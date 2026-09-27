# GDT — Serveur

Serveur backend de l'application **GDT**, une application web de gestion de notes organisées par espaces, réalisée dans le cadre d'un test technique.

Il expose une API REST développée en Go, responsable de la logique métier, de l'accès aux données, de l'authentification, de la validation des données et du contrôle d'accès.

> 🚧 Projet en cours de développement : l'authentification est disponible, la gestion des espaces et des notes est en cours.

## Stack technique

- **Langage :** Go, avec le routeur HTTP de la bibliothèque standard (`net/http`)
- **Base de données :** PostgreSQL (image Docker officielle), via le driver [pgx](https://github.com/jackc/pgx)
- **Migrations :** [golang-migrate](https://github.com/golang-migrate/migrate), fichiers SQL embarqués dans le binaire et appliqués au démarrage
- **Authentification :** tokens JWT signés en HS256 ([golang-jwt](https://github.com/golang-jwt/jwt)), mots de passe hachés avec bcrypt
- **Conteneurisation :** Docker Compose (base de données)

## Prérequis

- [Go](https://go.dev/dl/) (version indiquée dans `go.mod`)
- [Docker](https://docs.docker.com/get-docker/) et Docker Compose
- Git

## Installation

```bash
git clone https://github.com/CookieG77/AppGDT-Server.git
cd AppGDT-Server
```

Copier le modèle de configuration :

```bash
# Linux / macOS / Git Bash
cp .env.example .env
```

```powershell
# Windows (PowerShell)
Copy-Item .env.example .env
```

Puis renseigner au minimum les variables obligatoires (voir ci-dessous), en particulier `JWT_SECRET`.

## Configuration

La configuration est lue depuis les variables d'environnement. Un fichier `.env` placé à la racine est chargé automatiquement s'il existe ; les variables déjà définies dans l'environnement restent prioritaires. Le fichier `.env` contient des secrets : il n'est pas versionné.

### Serveur

| Variable  | Obligatoire | Défaut      | Description                    |
|-----------|-------------|-------------|--------------------------------|
| `ADDRESS` | Non         | `localhost` | Adresse d'écoute de l'API      |
| `PORT`    | Non         | `8080`      | Port d'écoute de l'API         |

### Base de données

Ces variables sont partagées entre Docker Compose, qui crée la base, et le serveur, qui s'y connecte.

| Variable            | Obligatoire | Défaut      | Description                          |
|---------------------|-------------|-------------|--------------------------------------|
| `POSTGRES_USER`     | Oui         |             | Utilisateur de la base de données    |
| `POSTGRES_PASSWORD` | Oui         |             | Mot de passe de la base de données   |
| `POSTGRES_DB`       | Non         | `gdt`       | Nom de la base de données            |
| `POSTGRES_HOST`     | Non         | `localhost` | Hôte de la base de données           |
| `DB_PORT`           | Non         | `5432`      | Port de la base de données           |

### Authentification

| Variable      | Obligatoire | Défaut | Description                                                          |
|---------------|-------------|--------|----------------------------------------------------------------------|
| `JWT_SECRET`  | Oui         |        | Clé de signature des tokens, entre 32 et 256 caractères, aléatoire   |
| `JWT_TTL`     | Non         | `1h`   | Durée de validité des tokens, entre `5m` et `24h` (ex. `30m`, `2h`)  |
| `BCRYPT_COST` | Non         | `12`   | Coût du hachage des mots de passe, entre `10` et `14`                |

Le serveur refuse de démarrer si une variable obligatoire est absente ou si une valeur sort des plages autorisées.

Pour générer une clé `JWT_SECRET` aléatoire :

```bash
# Linux / macOS / Git Bash (inclus avec Git pour Windows)
openssl rand -base64 32
```

```powershell
# PowerShell 7
[Convert]::ToBase64String([System.Security.Cryptography.RandomNumberGenerator]::GetBytes(32))
```

## Lancement

### 1. Démarrer la base de données

```bash
docker compose up -d
```

La base PostgreSQL est exposée uniquement en local, et ses données sont conservées dans un volume Docker. La commande `docker compose ps` permet de vérifier que le conteneur est à l'état `healthy`.

Pour arrêter la base : `docker compose down`. Pour la supprimer avec toutes ses données : `docker compose down -v`.

### 2. Démarrer le serveur

```bash
go run ./cmd/server
```

Au démarrage, le serveur applique automatiquement les migrations manquantes, se connecte à la base, puis écoute sur `http://localhost:8080` (selon `ADDRESS` et `PORT`). Les logs sont écrits au format JSON sur la sortie standard.

Pour vérifier que l'API répond :

```bash
curl http://localhost:8080/health
```

Le serveur s'arrête proprement avec `Ctrl+C` : les requêtes en cours ont jusqu'à 10 secondes pour se terminer.

## Migrations

Le schéma de la base est défini par les fichiers du dossier `migrations/`, appliqués dans l'ordre de leur numéro. Ils sont intégrés au binaire à la compilation et appliqués automatiquement au démarrage du serveur : seules les migrations pas encore appliquées sont exécutées.

| Migration              | Contenu                                                        |
|------------------------|----------------------------------------------------------------|
| `000001_create_users`  | Table `users`, fonction `set_updated_at` et son trigger        |
| `000002_create_spaces` | Table `spaces`, liée à `users`                                 |
| `000003_create_notes`  | Type énuméré `note_status` et table `notes`, liée à `spaces`   |

Chaque migration possède un fichier `up` (application) et un fichier `down` (annulation).

Pour vérifier la version appliquée :

```bash
docker compose exec postgres psql -U gdt -d gdt -c "SELECT * FROM schema_migrations;"
```

(Remplacer `gdt` par les valeurs de `POSTGRES_USER` et `POSTGRES_DB`.)

## API

Le contrat complet de l'API est décrit au format OpenAPI 3.1 dans [`api/openapi.yaml`](api/openapi.yaml). Pour le consulter sous forme de documentation interactive, coller son contenu dans [Swagger Editor](https://editor.swagger.io/).

Les routes protégées attendent un en-tête `Authorization: Bearer <token>`, le token étant obtenu via `POST /auth/login`.

| Méthode | Route            | Authentification | Description                          | État          |
|---------|------------------|------------------|--------------------------------------|---------------|
| `GET`   | `/health`        | Non              | Vérifie que l'API répond             | Disponible    |
| `POST`  | `/auth/register` | Non              | Crée un compte                       | Disponible    |
| `POST`  | `/auth/login`    | Non              | Renvoie un token JWT                 | Disponible    |
| `GET`   | `/users/me`      | Oui              | Profil de l'utilisateur connecté     | Disponible    |
| —       | `/spaces/...`    | Oui              | Gestion des espaces                  | En cours      |
| —       | `/notes/...`     | Oui              | Gestion des notes                    | En cours      |

## Tests

Une collection Postman couvre l'API d'authentification, y compris les cas d'erreur et plusieurs tentatives d'attaque (token falsifié, algorithme `none`, énumération des comptes, champs non autorisés) : [`api/postman/gdt-api-auth.postman_collection.json`](api/postman/gdt-api-auth.postman_collection.json).

Pour l'exécuter : l'importer dans Postman, puis lancer la collection complète avec le *Collection Runner*, **dans l'ordre**, serveur et base démarrés. L'adresse de l'API se règle dans la variable de collection `baseUrl`.

## Structure du projet

```
.
├── api/                  # Contrat OpenAPI et collection Postman
├── cmd/server/           # Point d'entrée : assemblage des dépendances et cycle de vie du serveur
├── docs/                 # Journal des choix techniques, schéma de la base
├── internal/
│   ├── auth/             # Hachage des mots de passe (bcrypt) et gestion des JWT
│   ├── config/           # Lecture et validation de la configuration
│   ├── database/         # Connexion à PostgreSQL et application des migrations
│   ├── domain/           # Entités métier et erreurs partagées entre les couches
│   ├── handler/          # Couche HTTP : lecture des requêtes, écriture des réponses
│   ├── httpjson/         # Lecture et écriture du JSON, format d'erreur commun
│   ├── middleware/       # Middleware d'authentification
│   ├── repository/       # Accès aux données (requêtes SQL)
│   ├── server/           # Déclaration des routes et configuration du serveur HTTP
│   └── service/          # Logique métier et validation
├── migrations/           # Migrations SQL, embarquées dans le binaire
├── docker-compose.yml
└── .env.example
```

Chaque requête traverse les couches dans cet ordre : `middleware` → `handler` → `service` → `repository` → PostgreSQL. Chaque couche ne dépend que de celle située en dessous, et le contrôle d'accès est garanti au niveau des requêtes SQL, qui filtrent systématiquement sur l'utilisateur authentifié.

## Documentation

- [Journal des choix techniques](docs/NOTES.md)
- [Schéma relationnel de la base de données](docs/database-schema.drawio.svg)
- [Contrat de l'API (OpenAPI)](api/openapi.yaml)

<!--
TODO : sections à ajouter / compléter
- Comptes de démonstration
- Lien vers le dépôt client
- Mise à jour du tableau des routes (espaces et notes)
-->
