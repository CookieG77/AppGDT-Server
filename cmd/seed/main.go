// Command seed creates demonstration accounts, with example spaces and notes,
// so that the application can be tried without registering first.
//
// It uses the same configuration (.env) and the same business rules as the
// server: accounts go through the registration service (validation, bcrypt
// hashing), spaces and notes through their services.
//
// Usage:
//
//	go run ./cmd/seed                     # creates the demo accounts, skips the existing ones
//	go run ./cmd/seed -reset              # deletes then recreates the demo accounts
//	go run ./cmd/seed -email a@b.fr -username Alice -password 'motdepasse'
//	                                      # creates one custom account with the example data
//
// The demo accounts have a publicly known password: they are meant for local
// or demonstration environments only, never for production.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/CookieG77/AppGDT-Server/internal/auth"
	"github.com/CookieG77/AppGDT-Server/internal/config"
	"github.com/CookieG77/AppGDT-Server/internal/database"
	"github.com/CookieG77/AppGDT-Server/internal/domain"
	"github.com/CookieG77/AppGDT-Server/internal/logging"
	"github.com/CookieG77/AppGDT-Server/internal/ratelimit"
	"github.com/CookieG77/AppGDT-Server/internal/repository"
	"github.com/CookieG77/AppGDT-Server/internal/service"
)

// options are the command line flags.
type options struct {
	email    string
	username string
	password string
	empty    bool
	reset    bool
	verbose  bool
}

func main() {
	opts := parseFlags()

	// The services log security events (user_registered, account_deleted…):
	// with -v they are written as JSON on stderr, so that stdout only shows
	// the summary. Without -v, only errors are logged.
	level := slog.LevelError
	if opts.verbose {
		level = slog.LevelInfo
	}
	slog.SetDefault(slog.New(logging.NewContextHandler(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level}))))

	if err := run(opts, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "Erreur :", err)
		os.Exit(1)
	}
}

func parseFlags() options {
	var opts options
	flag.StringVar(&opts.email, "email", "", "email d'un compte personnalisé à créer (avec -username et -password)")
	flag.StringVar(&opts.username, "username", "", "nom ou pseudo du compte personnalisé")
	flag.StringVar(&opts.password, "password", "", "mot de passe du compte personnalisé")
	flag.BoolVar(&opts.empty, "empty", false, "créer les comptes sans espaces ni notes d'exemple")
	flag.BoolVar(&opts.reset, "reset", false, "supprimer d'abord les comptes existants (avec toutes leurs données) puis les recréer")
	flag.BoolVar(&opts.verbose, "v", false, "afficher aussi les logs de sécurité des services (sur stderr)")
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintln(out, "Crée des comptes de démonstration GDT, avec des espaces et des notes d'exemple.")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Utilisation :")
		fmt.Fprintln(out, "  go run ./cmd/seed [-reset] [-empty]")
		fmt.Fprintln(out, "  go run ./cmd/seed -email EMAIL -username NOM -password MOTDEPASSE [-reset] [-empty]")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Options :")
		flag.PrintDefaults()
	}
	flag.Parse()
	return opts
}

func run(opts options, out io.Writer) error {
	accounts, err := accountsToCreate(opts)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	dbURL := database.BuildDatabaseURL(cfg.DatabaseCfg)
	if err := database.Migrate(dbURL); err != nil {
		return err
	}
	pool, err := database.Connect(ctx, dbURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	s, err := newSeeder(cfg, repository.NewUserRepository(pool), repository.NewSpaceRepository(pool), repository.NewNoteRepository(pool))
	if err != nil {
		return err
	}

	if opts.email == "" {
		fmt.Fprintln(out, "Comptes de démonstration GDT")
	} else {
		fmt.Fprintln(out, "Compte GDT")
	}
	fmt.Fprintln(out)
	for _, account := range accounts {
		if !opts.empty && len(account.Spaces) == 0 {
			account.Spaces = exampleSpaces
		}
		if opts.empty {
			account.Spaces = nil
		}

		result, err := s.seed(ctx, account, opts.reset)
		if err != nil {
			return fmt.Errorf("compte %s : %w", account.Email, err)
		}
		password := account.Password
		if opts.email != "" {
			password = "(celui fourni)" // not echoed: it may be a real password
		}
		fmt.Fprintf(out, "  %-9s %-24s mot de passe : %-14s %s\n", result.status, account.Email, password, result.details)
	}

	if opts.email == "" {
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Ces comptes ont un mot de passe connu : ne jamais les créer en production.")
	}
	return nil
}

// accountsToCreate returns the custom account given by the flags, or the
// default demo accounts when no custom account is requested.
func accountsToCreate(opts options) ([]demoAccount, error) {
	if opts.email == "" && opts.username == "" && opts.password == "" {
		return demoAccounts, nil
	}
	if opts.email == "" || opts.username == "" || opts.password == "" {
		return nil, errors.New("-email, -username et -password doivent être fournis ensemble (voir -h)")
	}
	return []demoAccount{{Email: opts.email, Username: opts.username, Password: opts.password}}, nil
}

// seeder creates accounts and their content through the services.
type seeder struct {
	users  *repository.UserRepository
	auth   *service.AuthService
	spaces *service.SpaceService
	notes  *service.NoteService
}

func newSeeder(cfg *config.Config, users *repository.UserRepository, spaces *repository.SpaceRepository, notes *repository.NoteRepository) (*seeder, error) {
	// The token manager and the login limiters are required by the service
	// but never used here: no one logs in through this command.
	limiters := service.LoginLimiters{
		ByEmail: ratelimit.New(cfg.LoginCfg.MaxPerEmail, cfg.LoginCfg.Window),
		ByIP:    ratelimit.New(cfg.LoginCfg.MaxPerIP, cfg.LoginCfg.Window),
	}
	authService, err := service.NewAuthService(users, auth.NewPasswordHasher(cfg.HashingCfg.Cost), auth.NewTokenManager(cfg.TokenCfg.Secret, cfg.TokenCfg.TTL), limiters)
	if err != nil {
		return nil, err
	}
	return &seeder{
		users:  users,
		auth:   authService,
		spaces: service.NewSpaceService(spaces),
		notes:  service.NewNoteService(notes, spaces),
	}, nil
}

type seedResult struct {
	status  string
	details string
}

// seed creates one account and its example content. An existing account is
// left untouched, unless reset is true: it is then deleted with all its data
// and created again.
func (s *seeder) seed(ctx context.Context, account demoAccount, reset bool) (seedResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Same normalization as the registration service, so that the lookup
	// finds the account whatever the case used on the command line
	email := strings.ToLower(strings.TrimSpace(account.Email))

	recreated := false
	existing, err := s.users.GetByEmail(ctx, email)
	switch {
	case err == nil && !reset:
		return seedResult{status: "existant", details: "(inchangé, utiliser -reset pour le recréer)"}, nil
	case err == nil && reset:
		// The spaces and notes are deleted with the user (ON DELETE CASCADE)
		if err := s.users.Delete(ctx, existing.ID); err != nil {
			return seedResult{}, fmt.Errorf("suppression de l'ancien compte : %w", err)
		}
		recreated = true
	case !errors.Is(err, domain.ErrNotFound):
		return seedResult{}, err
	}

	user, err := s.auth.Register(ctx, email, account.Username, account.Password)
	if err != nil {
		return seedResult{}, describe(err)
	}

	noteCount := 0
	for _, sp := range account.Spaces {
		space, err := s.spaces.Create(ctx, user.ID, sp.Name, sp.Description)
		if err != nil {
			return seedResult{}, fmt.Errorf("espace %q : %w", sp.Name, describe(err))
		}
		for _, n := range sp.Notes {
			status := n.Status
			if _, err := s.notes.Create(ctx, user.ID, space.ID, service.NoteInput{Title: n.Title, Content: n.Content, Status: &status}); err != nil {
				return seedResult{}, fmt.Errorf("note %q : %w", n.Title, describe(err))
			}
			noteCount++
		}
	}

	status := "créé"
	if recreated {
		status = "recréé"
	}
	return seedResult{status: status, details: "(" + plural(len(account.Spaces), "espace") + ", " + plural(noteCount, "note") + ")"}, nil
}

// plural writes "1 note", "2 notes"…
func plural(n int, word string) string {
	if n > 1 {
		word += "s"
	}
	return fmt.Sprintf("%d %s", n, word)
}

// describe turns a validation error into a readable message.
func describe(err error) error {
	var vErr *domain.ValidationError
	if errors.As(err, &vErr) {
		msg := "données invalides :"
		for _, f := range vErr.Fields {
			msg += fmt.Sprintf(" %s (%s)", f.Message, f.Field)
		}
		return errors.New(msg)
	}
	return err
}
