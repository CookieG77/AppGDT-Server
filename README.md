# GDT — Serveur

Serveur backend de l'application **GDT**, une application web de gestion de notes organisées par espaces, réalisée dans le cadre d'un test technique.

Il exposera une API REST développée en Go, responsable de la logique métier, de l'accès aux données, de l'authentification, de la validation des données et du contrôle d'accès.

> 🚧 Projet en cours de développement.

## Stack technique

- **Base de données :** PostgreSQL (image Docker officielle)
- **Conteneurisation :** Docker Compose
- **Schéma de la base :** fichiers de migration SQL versionnés (`up` / `down`)

## Prérequis

- [Docker](https://docs.docker.com/get-docker/) et Docker Compose
- Git

## Configuration

Copier le modèle de configuration, puis adapter les valeurs si besoin :

```bash
cp .env.example .env
```

| Variable            | Description                           |
|---------------------|---------------------------------------|
| `POSTGRES_USER`     | Utilisateur de la base de données     |
| `POSTGRES_PASSWORD` | Mot de passe de la base de données    |
| `POSTGRES_DB`       | Nom de la base de données             |

Le fichier `.env` contient des secrets : il n'est pas versionné.

## Base de données

### Démarrage

```bash
docker compose up -d
```

La base PostgreSQL est exposée uniquement en local sur le port `5432`, et ses données sont conservées dans un volume Docker. La commande `docker compose ps` permet de vérifier que le conteneur est à l'état `healthy`.

Pour arrêter la base : `docker compose down`. Pour la supprimer avec toutes ses données : `docker compose down -v`.

### Migrations

Le schéma de la base est défini par les fichiers du dossier `migrations/`, appliqués dans l'ordre de leur numéro :

| Migration | Contenu |
|-----------|---------|
| `000001_create_users`  | Table `users`, fonction `set_updated_at` et son trigger |
| `000002_create_spaces` | Table `spaces`, liée à `users` |
| `000003_create_notes`  | Type énuméré `note_status` et table `notes`, liée à `spaces` |

Chaque migration possède un fichier `up` (application) et un fichier `down` (annulation).

## Documentation

- [Journal des choix techniques](docs/NOTES.md)
- [Schéma relationnel de la base de données](docs/database-schema.drawio.svg)

<!--
TODO : Section à ajouter/Compléter
- Installation (clonage, dépendances Go)
- Lancement du serveur
- Comptes de démonstration
- Documentation de l'API
- Structure du projet
- Lien vers le dépôt client
-->