# GDT — Serveur

Serveur backend de l'application **GDT**, une application web de gestion de notes organisées par espaces, réalisée dans le cadre d'un test technique.

Il expose une API REST développée en Go, responsable de la logique métier, de l'accès aux données, de l'authentification, de la validation des données et du contrôle d'accès.

> 🚧 Projet en cours de développement : l'API (authentification, espaces et notes) est fonctionnelle et couverte par une collection de tests Postman. Les comptes de démonstration et plusieurs améliorations (robustesse, sécurité) restent à ajouter.

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

### Limitation des tentatives de connexion

| Variable                       | Obligatoire | Défaut | Description                                                                   |
|--------------------------------|-------------|--------|-------------------------------------------------------------------------------|
| `LOGIN_MAX_FAILURES_PER_EMAIL` | Non         | `5`    | Échecs de connexion autorisés par email avant blocage, entre `3` et `20`      |
| `LOGIN_MAX_FAILURES_PER_IP`    | Non         | `50`   | Échecs de connexion autorisés par IP avant blocage, entre `10` et `1000`      |
| `LOGIN_FAILURE_WINDOW`         | Non         | `15m`  | Fenêtre de comptage des échecs et durée du blocage, entre `1m` et `24h`       |

Voir [Limitation des tentatives de connexion](#limitation-des-tentatives-de-connexion) pour le fonctionnement.

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

## Journalisation

Les logs sont écrits au format JSON sur la sortie standard, une ligne par événement. Ils sont de trois types :

- **Requêtes** (`"msg": "request handled"`) : une ligne par requête avec la méthode, le chemin, le statut, la taille de la réponse, la durée et l'IP du client. Le niveau dépend du statut : `INFO` pour un succès, `WARN` pour une erreur client (4xx), `ERROR` pour une erreur serveur (5xx).
- **Événements de sécurité** (`"msg": "security event"`) : filtrables par leur attribut `event`, ils incluent toujours l'IP du client.
- **Actions métier** : création, modification et suppression des espaces et des notes, avec leur identifiant.

| Événement de sécurité   | Niveau | Déclencheur                                                              |
|-------------------------|--------|--------------------------------------------------------------------------|
| `user_registered`       | `INFO` | Création d'un compte                                                     |
| `registration_rejected` | `INFO` | Inscription avec une adresse déjà utilisée                               |
| `login_succeeded`       | `INFO` | Connexion réussie                                                        |
| `login_failed`          | `WARN` | Connexion refusée, avec la raison (email inconnu, mauvais mot de passe)  |
| `login_locked`          | `WARN` | Un email ou une IP atteint la limite d'échecs de connexion (`scope`)     |
| `login_blocked`         | `WARN` | Connexion refusée car l'email ou l'IP est bloqué (`scope`)               |
| `token_rejected`        | `WARN` | Token invalide, falsifié ou expiré                                       |
| `resource_not_found`    | `INFO` | Ressource inexistante ou appartenant à un autre utilisateur              |

Chaque requête reçoit un identifiant, renvoyé dans l'en-tête de réponse `X-Request-ID` et ajouté à tous les logs écrits pendant son traitement, avec l'identifiant de l'utilisateur connecté (`userID`). Un client peut transmettre son propre identifiant dans ce même en-tête pour suivre une requête de bout en bout ; il est réutilisé s'il est valide (64 caractères max, lettres, chiffres, `.`, `_` et `-`).

Les raisons précises d'un échec de connexion n'apparaissent que dans les logs : le client reçoit toujours la même erreur. Aucune donnée personnelle n'est journalisée (ni email, ni mot de passe, ni token, ni corps de requête) : les utilisateurs sont identifiés uniquement par leur identifiant.

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

Les routes protégées attendent un en-tête `Authorization: Bearer <token>`, le token étant obtenu via `POST /auth/login`. Les corps de requête et de réponse sont au format JSON.

### Routes

| Méthode  | Route                     | Authentification | Description                                   | Succès |
|----------|---------------------------|------------------|-----------------------------------------------|--------|
| `GET`    | `/health`                 | Non              | Vérifie que l'API répond                      | `200`  |
| `POST`   | `/auth/register`          | Non              | Crée un compte                                | `201`  |
| `POST`   | `/auth/login`             | Non              | Renvoie un token JWT                          | `200`  |
| `GET`    | `/users/me`               | Oui              | Profil de l'utilisateur connecté              | `200`  |
| `GET`    | `/spaces`                 | Oui              | Liste les espaces de l'utilisateur            | `200`  |
| `POST`   | `/spaces`                 | Oui              | Crée un espace                                | `201`  |
| `GET`    | `/spaces/{spaceId}`       | Oui              | Consulte un espace                            | `200`  |
| `PUT`    | `/spaces/{spaceId}`       | Oui              | Modifie le nom et la description d'un espace  | `200`  |
| `DELETE` | `/spaces/{spaceId}`       | Oui              | Supprime un espace et toutes ses notes        | `204`  |
| `GET`    | `/spaces/{spaceId}/notes` | Oui              | Liste les notes d'un espace                   | `200`  |
| `POST`   | `/spaces/{spaceId}/notes` | Oui              | Crée une note dans un espace                  | `201`  |
| `GET`    | `/notes/{noteId}`         | Oui              | Consulte une note                             | `200`  |
| `PUT`    | `/notes/{noteId}`         | Oui              | Modifie le titre, le contenu et l'état        | `200`  |
| `DELETE` | `/notes/{noteId}`         | Oui              | Supprime une note                             | `204`  |

Les listes sont triées de la plus récente à la plus ancienne. Les routes `PUT` remplacent l'intégralité de la ressource : un champ facultatif omis reprend sa valeur par défaut.

### Règles de validation

| Ressource   | Champ         | Règle                                                                    |
|-------------|---------------|--------------------------------------------------------------------------|
| Utilisateur | `email`       | Obligatoire, format valide, 254 caractères max, insensible à la casse    |
| Utilisateur | `username`    | Obligatoire, 50 caractères max                                           |
| Utilisateur | `password`    | 8 caractères min, 72 octets max (limite de bcrypt)                       |
| Espace      | `name`        | Obligatoire, 100 caractères max                                          |
| Espace      | `description` | Facultative (vide par défaut), 1000 caractères max                       |
| Note        | `title`       | Obligatoire, 200 caractères max                                          |
| Note        | `content`     | Facultatif (vide par défaut), 50000 caractères max, conservé tel quel    |
| Note        | `status`      | `todo` (défaut), `in_progress` ou `done`                                 |

Les espaces superflus en début et en fin de champ sont retirés, sauf pour le contenu des notes. Les longueurs sont comptées en caractères et non en octets. Tout champ non prévu par le contrat est refusé.

### Contrôle d'accès

Un utilisateur n'accède qu'à ses propres espaces et notes. Une ressource appartenant à un autre utilisateur est traitée comme inexistante : la réponse est un `404` identique à celui d'une ressource qui n'existe pas, afin de ne pas révéler son existence. Une note est toujours rattachée à l'espace indiqué dans le chemin lors de sa création, et ne peut pas être déplacée vers un autre espace.

### Limitation des tentatives de connexion

Pour freiner la recherche de mots de passe par force brute, les connexions échouées sont comptées de deux façons :

- **par email** : protège chaque compte contre une attaque ciblée. Les échecs sont comptés même pour un email sans compte, pour qu'un blocage ne révèle pas si l'adresse est inscrite ;
- **par IP** : limite un même client qui essaierait des mots de passe sur de nombreux comptes. La limite est plus haute, car plusieurs utilisateurs peuvent partager une IP (réseau d'entreprise ou d'école).

Une fois la limite atteinte, la connexion est refusée avec un `429 TOO_MANY_ATTEMPTS`, **même avec le bon mot de passe**, pendant toute la durée de la fenêtre. L'en-tête `Retry-After` indique le nombre de secondes à attendre. Une connexion réussie remet à zéro le compteur de l'email, mais pas celui de l'IP : sinon, un attaquant pourrait l'effacer en se connectant à son propre compte entre deux essais.

Les compteurs sont gardés en mémoire : ils sont perdus au redémarrage du serveur et ne seraient pas partagés entre plusieurs instances. C'est suffisant pour un serveur unique ; plusieurs instances nécessiteraient un stockage partagé (base de données, Redis). Le blocage par email permet aussi à un tiers de bloquer temporairement un compte en échouant volontairement : c'est le compromis habituel de ce mécanisme, limité par la durée de la fenêtre.

### Format des erreurs

Toutes les erreurs de l'API suivent le même format, y compris pour une route inexistante ou une méthode non autorisée :

```json
{
  "code": "VALIDATION_ERROR",
  "message": "Les données envoyées sont invalides.",
  "details": [
    { "field": "name", "message": "Le nom est obligatoire." }
  ]
}
```

Le champ `details` n'est présent que pour les erreurs de validation, et liste tous les champs invalides en une seule réponse.

| Statut | Code                  | Cas                                                                     |
|--------|-----------------------|-------------------------------------------------------------------------|
| `400`  | `VALIDATION_ERROR`    | Un ou plusieurs champs ne respectent pas les règles de validation      |
| `400`  | `INVALID_JSON`        | Corps vide, mal formé, mauvais type de champ ou champ non autorisé     |
| `400`  | `INVALID_ID`          | Identifiant du chemin qui n'est pas un entier positif                  |
| `401`  | `UNAUTHORIZED`        | Token absent, invalide ou expiré                                       |
| `401`  | `INVALID_CREDENTIALS` | Email ou mot de passe incorrect                                        |
| `404`  | `NOT_FOUND`           | Ressource inexistante ou appartenant à un autre utilisateur            |
| `404`  | `ROUTE_NOT_FOUND`     | Aucune route ne correspond au chemin demandé                           |
| `405`  | `METHOD_NOT_ALLOWED`  | Méthode non autorisée pour ce chemin (l'en-tête `Allow` liste les méthodes acceptées) |
| `409`  | `EMAIL_ALREADY_USED`  | Adresse email déjà utilisée                                            |
| `429`  | `TOO_MANY_ATTEMPTS`   | Trop de connexions échouées (l'en-tête `Retry-After` indique l'attente) |
| `500`  | `INTERNAL_ERROR`      | Erreur inattendue, y compris un panic dans un handler (le détail est journalisé, jamais renvoyé au client) |

## Tests

L'API est couverte par une collection Postman de plus de 500 assertions, rangée dans [`api/postman/`](api/postman/). Elle est organisée en dossiers numérotés :

| Dossiers | Domaine        | Contenu                                                                                               |
|----------|----------------|-------------------------------------------------------------------------------------------------------|
| 0 à 3    | Authentification | Inscription, connexion, profil, tentatives d'attaque (token falsifié, algorithme `none`, énumération des comptes) |
| 4 à 8    | Espaces        | Création, consultation, modification, isolation entre utilisateurs, suppression                     |
| 9 à 13   | Notes          | Création, consultation, modification, isolation entre utilisateurs, suppression en cascade         |
| 14       | Connexion      | Blocage après trop d'échecs pour un email, même avec le bon mot de passe, sans bloquer les autres comptes |

Chaque domaine vérifie les cas nominaux, les limites exactes de validation, les corps et identifiants invalides, l'absence de token, et le fait qu'un second utilisateur ne peut ni voir, ni modifier, ni supprimer les ressources du premier.

Pour l'exécuter :

1. Démarrer la base de données et le serveur (voir [Lancement](#lancement)).
2. Dans Postman, ouvrir le dossier `api/postman/` : la collection *GDT API - Authentification* apparaît dans la vue locale (*Local View*).
3. Lancer la collection **complète** avec le *Collection Runner*, **dans l'ordre** : les dossiers réutilisent les variables créées par les précédents (tokens, identifiants).

L'adresse de l'API se règle dans la variable de collection `baseUrl` (`http://localhost:8080` par défaut). Chaque exécution crée des comptes avec des adresses uniques et supprime les espaces qu'elle a créés : la collection peut être relancée sans réinitialiser la base.

Chaque exécution compte 7 connexions échouées pour l'IP du poste de test. Avec la limite par défaut (50 par quart d'heure), la collection peut donc être lancée 7 fois par quart d'heure ; au-delà, redémarrer le serveur remet les compteurs à zéro. Le dossier 14 suppose la limite par email par défaut (`5`) : si `LOGIN_MAX_FAILURES_PER_EMAIL` est modifiée, mettre à jour la variable de collection `loginMaxFailuresPerEmail`.

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
│   ├── logging/          # Contexte des logs (identifiant de requête, utilisateur) et événements de sécurité
│   ├── middleware/       # Middlewares d'authentification, de journalisation et de récupération des panics
│   ├── ratelimit/        # Comptage des échecs et blocage (limitation des tentatives de connexion)
│   ├── repository/       # Accès aux données (requêtes SQL)
│   ├── server/           # Déclaration des routes, erreurs JSON des routes inconnues, configuration du serveur HTTP
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
-->
