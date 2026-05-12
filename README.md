# captive-portal-admin

Panel admin local en Go pour gérer les tickets WiFi temporaires d'un camping.

Le projet vise une interface serveur-side simple, accessible depuis le réseau
admin local, avec PostgreSQL comme base cible. Les futures étapes couvriront les
emplacements, la synchronisation FreeRADIUS, l'authentification admin et
l'impression ou l'affichage des tickets.

Le projet voisin `captive-portal` separe deja les donnees applicatives du
portail, les tables techniques FreeRADIUS et les logs legaux. Ce panel admin
gardera la meme separation : donnees metier admin d'un cote, synchronisation
vers la base RADIUS de l'autre, sans modifier directement les logs legaux du
portail client.

## Lancement

```sh
APP_ADDR="127.0.0.1:8081" \
DATABASE_URL="postgres://admin_user:admin_password@127.0.0.1:5432/admin?sslmode=disable" \
RADIUS_DATABASE_URL="postgres://radius_user:radius_dev_password@127.0.0.1:5432/radius?sslmode=disable" \
  go run ./cmd/admin-panel
```

Ou via Make :

```sh
APP_ADDR="127.0.0.1:8081" \
DATABASE_URL="postgres://admin_user:admin_password@127.0.0.1:5432/admin?sslmode=disable" \
RADIUS_DATABASE_URL="postgres://radius_user:radius_dev_password@127.0.0.1:5432/radius?sslmode=disable" \
  make run
make build
make test
make fmt
```

## Docker

Build local :

```sh
docker build -t captive-portal-admin:dev .
```

Run local, avec PostgreSQL disponible sur la machine :

```sh
docker run --rm \
  -p 8081:8080 \
  -e APP_ADDR=0.0.0.0:8080 \
  -e DATABASE_URL="postgres://admin_user:admin_password@host.docker.internal:5432/admin?sslmode=disable" \
  -e RADIUS_DATABASE_URL="postgres://radius_user:radius_dev_password@host.docker.internal:5432/radius?sslmode=disable" \
  captive-portal-admin:dev
```

Dans le labo `../test-env`, le panel admin est expose sur :

```text
http://127.0.0.1:8081
```

Le conteneur ecoute en interne sur `0.0.0.0:8080` via `APP_ADDR`, avec le
mapping Compose `8081:8080`. La base metier admin utilisee est `admin` :

```text
postgres://admin_user:admin_password@postgres:5432/admin?sslmode=disable
```

La synchronisation FreeRADIUS utilise une connexion separee vers `radius` :

```text
postgres://radius_user:radius_dev_password@postgres:5432/radius?sslmode=disable
```

Demarrage depuis `../test-env` :

```sh
docker compose up -d --build
curl -i http://127.0.0.1:8081/healthz
```

Image GHCR publiee par la CI :

```sh
docker pull ghcr.io/justarandombaddev/captive-portal-admin:latest
docker pull ghcr.io/justarandombaddev/captive-portal-admin:sha-<commit>
docker pull ghcr.io/justarandombaddev/captive-portal-admin:vX.Y.Z
```

## Variables d'environnement

| Variable              | Description                                 | Defaut  |
| --------------------- | ------------------------------------------- | ------- |
| `APP_ADDR`            | Adresse d'ecoute HTTP                       | `:8080` |
| `DATABASE_URL`        | URL PostgreSQL de la base metier admin      | requis  |
| `RADIUS_DATABASE_URL` | URL PostgreSQL de la base FreeRADIUS        | requis  |
| `SESSION_SECRET`      | Secret de session pour la future auth admin | vide    |

## Routes

- `GET /` : dashboard placeholder.
- `GET /tickets` : liste des tickets WiFi.
- `GET /tickets/new` : formulaire de creation d'un ticket.
- `POST /tickets` : creation d'un ticket temporaire.
- `POST /tickets/{id}/revoke` : revocation d'un ticket.
- `GET /pitches` : liste des emplacements.
- `GET /pitches/new` : formulaire de creation d'un emplacement.
- `POST /pitches` : creation d'un emplacement.
- `POST /pitches/{id}/disable` : desactivation d'un emplacement.
- `POST /pitches/{id}/enable` : reactivation d'un emplacement.
- `GET /healthz` : verifie PostgreSQL et retourne `OK`.

## Organisation backend

- `cmd/admin-panel` : point d'entree.
- `internal/app` : assemblage de l'application.
- `internal/config` : configuration par variables d'environnement.
- `internal/http` : routeur et handlers HTTP.
- `internal/database` : connexion PostgreSQL via `pgxpool`.
- `internal/tickets` : service metier tickets WiFi.
- `internal/pitches` : service metier emplacements.
- `internal/radius` : service de synchronisation FreeRADIUS futur.
- `internal/adminauth` : service d'authentification admin futur.
- `internal/templates` : vues HTML serveur-side.

## Modeles metier

Le panel admin gere deux agregats principaux, sans ORM lourd et avec des
repositories PostgreSQL explicites.

`Ticket` represente un acces WiFi temporaire :

- identifiant UUID
- username unique
- mot de passe temporaire en clair pour impression et synchronisation RADIUS
- emplacement associe
- statut `active`, `expired` ou `revoked`
- dates de validite
- informations de creation et de revocation
- horodatage de synchronisation FreeRADIUS futur

`Pitch` represente un emplacement du camping :

- identifiant UUID
- code ou numero unique
- libelle optionnel
- activation ou desactivation
- timestamps de creation et mise a jour

Les packages `tickets` et `pitches` exposent chacun une interface `Repository`
et une implementation PostgreSQL explicite basee sur `pgxpool`. `internal/app`
injecte ces repositories SQL au demarrage. Le serveur refuse de demarrer si
`DATABASE_URL`, `RADIUS_DATABASE_URL` ou PostgreSQL sont indisponibles.

## Synchronisation FreeRADIUS

`internal/radius` expose une interface `Syncer` pour isoler le flux :

```text
admin panel -> RadiusSync -> FreeRADIUS DB
```

Le service tickets appelle cette interface apres creation, expiration ou
revocation d'un ticket. L'implementation PostgreSQL ouvre une connexion dediee
vers `radius` via `RADIUS_DATABASE_URL` et ne reutilise jamais la connexion
`admin`.

Creation d'un ticket :

- cree ou reactive l'entree `radius_users` ;
- remplace les check items `radcheck` du ticket ;
- ajoute `Cleartext-Password := <mot de passe>` ;
- ajoute `Expiration := <date de fin>` pour que FreeRADIUS refuse le ticket
  apres `valid_until`.

Revocation ou expiration :

- supprime les credentials `radcheck` ;
- supprime les reponses/groupes eventuels du username ;
- desactive `radius_users`.

Une erreur de synchronisation ne rollback pas la donnee metier admin ; elle est
journalisee pour permettre une reprise ou une sync asynchrone plus tard. Le
champ `wifi_tickets.radius_synced_at` est mis a jour uniquement apres une sync
reussie. Les logs legaux restent geres par `captive-portal`.

## Migrations PostgreSQL

Les migrations dans `migrations/` concernent uniquement la base metier du panel
admin (`admin`). La migration initiale cree :

- `admin_users` : placeholder minimal pour la future authentification admin.
- `pitches` : emplacements du camping.
- `wifi_tickets` : tickets WiFi temporaires lies aux emplacements.

FreeRADIUS conserve ses propres tables techniques dans la base RADIUS
(`radcheck`, `radreply`, `radacct`, `radpostauth`, etc.). Les logs legaux de
connexion restent geres par le projet voisin `captive-portal` et sa base
applicative. Le panel admin ne cree pas et ne purge pas ces logs.
