package main

import "github.com/CookieG77/AppGDT-Server/internal/domain"

// demoPassword is shared by the demo accounts. It is public on purpose:
// these accounts must never exist on a production database.
const demoPassword = "Demo1234!"

type demoAccount struct {
	Email    string
	Username string
	Password string
	Spaces   []demoSpace
}

type demoSpace struct {
	Name        string
	Description string
	Notes       []demoNote
}

type demoNote struct {
	Title   string
	Content string
	Status  domain.NoteStatus
}

// demoAccounts are created when no custom account is given. Two accounts
// allow to check that a user never sees the spaces of another one.
var demoAccounts = []demoAccount{
	{Email: "demo@example.com", Username: "Démo", Password: demoPassword, Spaces: exampleSpaces},
	{Email: "camille@example.com", Username: "Camille", Password: demoPassword, Spaces: secondUserSpaces},
}

// exampleSpaces covers every feature: several spaces, the three statuses,
// empty content and Markdown formatting.
var exampleSpaces = []demoSpace{
	{
		Name:        "Devoirs",
		Description: "Travail scolaire à rendre et révisions.",
		Notes: []demoNote{
			{
				Title:   "Exercices de maths p.52",
				Content: "Faire les exercices **4 à 7**.\n\n- vérifier les unités\n- rendre lundi",
				Status:  domain.Done,
			},
			{
				Title: "Exposé d'histoire",
				Content: "## Plan\n\n1. Contexte\n2. Déroulement\n3. Conséquences\n\n" +
					"> Penser à citer les sources.\n\nSources : [Wikipédia](https://fr.wikipedia.org)",
				Status: domain.InProgress,
			},
			{
				Title:   "Réviser l'anglais",
				Content: "Vocabulaire des unités 3 et 4.",
				Status:  domain.Todo,
			},
		},
	},
	{
		Name:        "Jobs",
		Description: "Recherche d'emploi : candidatures, relances et entretiens.",
		Notes: []demoNote{
			{
				Title: "Candidatures envoyées",
				Content: "| Entreprise | Poste | Date |\n|---|---|---|\n" +
					"| Atelier Nord | Développeur Go | 12/09 |\n| Studio Sud | Développeur web | 18/09 |",
				Status: domain.InProgress,
			},
			{
				Title:   "Préparer l'entretien",
				Content: "- [x] Relire l'offre\n- [x] Préparer une présentation de 2 minutes\n- [ ] Lister trois questions à poser",
				Status:  domain.InProgress,
			},
			{
				Title:   "Mettre à jour le CV",
				Content: "",
				Status:  domain.Todo,
			},
		},
	},
	{
		Name:        "Personnel",
		Description: "",
		Notes: []demoNote{
			{
				Title:   "Idées de cadeaux",
				Content: "Un livre de cuisine, un jeu de société, une plante.",
				Status:  domain.Todo,
			},
		},
	},
	{
		Name:        "Projets",
		Description: "Un espace vide, pour voir l'affichage sans note.",
	},
}

var secondUserSpaces = []demoSpace{
	{
		Name:        "Sport",
		Description: "Programme d'entraînement.",
		Notes: []demoNote{
			{Title: "Course du dimanche", Content: "10 km en moins d'une heure.", Status: domain.Todo},
		},
	},
}
