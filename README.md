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
go run ./cmd/admin-panel
```

Ou via Make :

```sh
make run
make build
make test
make fmt
```

## Variables d'environnement

| Variable | Description | Defaut |
| --- | --- | --- |
| `APP_ADDR` | Adresse d'ecoute HTTP | `:8080` |
| `DATABASE_URL` | URL PostgreSQL cible | vide |
| `SESSION_SECRET` | Secret de session pour la future auth admin | vide |

## Routes

- `GET /` : dashboard placeholder.
- `GET /tickets` : page placeholder des tickets WiFi.
- `GET /pitches` : page placeholder des emplacements.
- `GET /healthz` : retourne `OK`.

## Organisation backend

- `cmd/admin-panel` : point d'entree.
- `internal/app` : assemblage de l'application.
- `internal/config` : configuration par variables d'environnement.
- `internal/http` : routeur et handlers HTTP.
- `internal/database` : placeholder du futur acces PostgreSQL.
- `internal/tickets` : service metier tickets WiFi.
- `internal/pitches` : service metier emplacements.
- `internal/radius` : service de synchronisation FreeRADIUS futur.
- `internal/adminauth` : service d'authentification admin futur.
- `internal/templates` : vues HTML serveur-side.

## Modeles metier

Le panel admin prepare deux agregats principaux, sans ORM lourd et sans requetes
SQL concretes pour l'instant.

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

Les packages `tickets` et `pitches` exposent chacun une interface `Repository`.
L'implementation actuelle est un repository memoire de developpement, injecte
par `internal/app`. Les futures migrations SQL devront creer des tables metier
admin separees des tables FreeRADIUS et des logs legaux du portail captif.
