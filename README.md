# eUprava — Saobraćajna policija (mikroservisna arhitektura)

Studentski projekat po uzoru na portal eUprava. Glavna tema je mikroservis
**Saobraćajna policija**, a sistem čini pet servisa koji komuniciraju preko
REST API-ja, svaki sa sopstvenom PostgreSQL bazom.

## Arhitektura

```
                        Angular (port 4200)
                              |
                        API Gateway (:8080)
                              |
     +------------+-----------+-----------+-------------+
     |            |                       |             |
 Citizen      Traffic Police          Payment      Notification
 (:8081)        (:8082)               (:8083)        (:8084)
     |            |                       |             |
 citizen_db   traffic_db             payment_db     notification_db
```

| Servis | Port | Odgovornost |
|---|---|---|
| API Gateway | 8080 | Jedina ulazna tačka: JWT validacija, autorizacija po ulozi, reverse proxy, CORS |
| Citizen | 8081 | Registracija/prijava (JWT), podaci o građanima |
| **Traffic Police** | 8082 | Vozači, prekršaji, novčane kazne, kazneni poeni, status dozvole |
| Payment | 8083 | Iniciranje i potvrda plaćanja kazni |
| Notification | 8084 | Elektronska obaveštenja građanima |

Komunikacija servis–servis (interna, van gateway-a):
- Traffic Police → Citizen: provera/dobavljanje podataka o građaninu
- Traffic Police → Notification: obaveštenje o novoj kazni / suspenziji dozvole
- Payment → Traffic Police: `GET /fines/{id}` (provera) i `PUT /fines/{id}/pay` (evidencija plaćanja)
- Payment → Notification: potvrda uplate

## Tehnologije

Go (Gin) · Angular 19 · PostgreSQL 16 (baza po servisu) · Docker Compose · JWT (HS256)

## Struktura repozitorijuma

```
eUprava/
├── docker-compose.yml
├── backend/
│   ├── api-gateway/            # ulazna tačka, JWT + role middleware, proxy
│   └── services/
│       ├── citizen/            # auth i podaci o građanima
│       ├── traffic-police/     # glavna tema projekta
│       ├── payment/            # plaćanja kazni
│       └── notification/       # obaveštenja
│           (svaki servis: cmd/main.go + internal/{handler,service,repository,model,client})
└── frontend/                   # Angular aplikacija
```

## Pokretanje

```bash
cp .env.example .env    
docker compose up --build
```

- Frontend: http://localhost:4200
- API Gateway: http://localhost:8080

**Podrazumevani nalozi** (seed-uju se automatski pri startu):
- policajac: `policija@euprava.rs` / `policija123`
- administrator: `admin@euprava.rs` / `admin123`

Građani se registruju kroz aplikaciju (uloga `citizen`); naloge policajaca
kreira administrator kroz stranicu "Korisnici".

### Pokretanje bez Dockera (razvoj)

Svaki servis je samostalan Go modul — potrebne su mu env promenljive
`DATABASE_URL`, `PORT` i URL-ovi servisa sa kojima komunicira (podrazumevane
vrednosti su u `cmd/main.go`). Frontend: `cd frontend && npm install && npm start`.

## Poslovna pravila (Traffic Police)

1. Unos prekršaja (`POST /violations`) automatski:
   - kreira novčanu kaznu sa rokom plaćanja 15 dana,
   - dodaje kaznene poene vozaču (iz šifarnika tipova prekršaja),
   - šalje obaveštenje građaninu preko Notification servisa.
2. Kada vozač dostigne **18 kaznenih poena**, dozvola se automatski
   suspenduje i građanin dobija obaveštenje.
3. Brisanje prekršaja vraća poene (i status dozvole ako padne ispod limita)
   i briše nevezanu kaznu; prekršaj sa **plaćenom** kaznom ne može da se obriše.
4. Plaćanjem kazne (Payment servis) kazna postaje plaćena, prekršaj rešen,
   a građanin dobija potvrdu.
5. Kao vozač može da se evidentira samo korisnik sa ulogom `citizen` —
   provera uloge ide kroz Citizen servis.
6. Open data: anonimna agregirana statistika prekršaja i šifarnik su javno
   dostupni bez prijave (stranica "Otvoreni podaci"); lični podaci su uvek iza JWT-a.

Šifarnik prekršaja: prekoračenje brzine (6p/20.000), crveno svetlo (8p/25.000),
nevezan pojas (2p/5.000), alkohol (14p/100.000), telefon (3p/10.000).

## REST API (kroz gateway, prefiks `/api`)

| Metoda i putanja | Uloga | Opis |
|---|---|---|
| `POST /api/auth/register` | javno | registracija građanina |
| `POST /api/auth/login` | javno | prijava, vraća JWT |
| `GET /api/citizens/me` | svi | sopstveni profil |
| `GET /api/citizens/{id}` | policajac | podaci o građaninu |
| `GET/POST /api/drivers` | policajac | evidencija vozača |
| `GET /api/drivers/{id}` | policajac | vozač + podaci o građaninu |
| `GET /api/drivers/{id}/violations` | policajac | istorija prekršaja vozača |
| `GET /api/penalty-points/{driverId}` | policajac | poeni i status dozvole |
| `GET/POST /api/violations`, `PUT/DELETE /api/violations/{id}` | policajac | CRUD prekršaja |
| `GET /api/fines` | policajac | sve kazne |
| `GET /api/me/driver`, `/api/me/violations`, `/api/me/fines` | građanin | sopstveni podaci |
| `POST /api/payments` | građanin | iniciranje plaćanja kazne |
| `POST /api/payments/{id}/confirm` | građanin | potvrda (simulacija uplate) |
| `GET /api/payments` | građanin | istorija plaćanja |
| `GET /api/notifications` | svi | sopstvena obaveštenja |
| `PUT /api/notifications/{id}/read` | svi | označi kao pročitano |
| `GET /api/users` | admin | pregled svih korisnika |
| `POST /api/users/officers` | admin | kreiranje naloga policajca |
| `GET /api/open-data/violation-stats` | **javno** | open data: anonimna statistika prekršaja |
| `GET /api/violation-types` | **javno** | open data: šifarnik prekršaja |

Interni endpointi (nisu dostupni spolja): `PUT /fines/{id}/pay` (Payment → Traffic
Police), `POST /notifications` (servisi → Notification).

## SOLID principi

- **S** — svaki servis ima slojeve `handler` (HTTP) → `service` (poslovna logika)
  → `repository` (baza), svaki sa jednom odgovornošću.
- **O/L** — implementacije repozitorijuma i klijenata su zamenljive
  (Postgres/HTTP u produkciji, mock u testovima) bez izmene service sloja.
- **I** — mali, fokusirani interfejsi: `DriverRepository`, `ViolationRepository`,
  `FineRepository`, `CitizenClient`, `NotificationClient`.
- **D** — service sloj zavisi od interfejsa; konkretne implementacije se
  ubrizgavaju u `cmd/main.go` (dependency injection).

Unit testovi poslovnih pravila sa mock-ovima:
`backend/services/traffic-police/internal/service/violation_service_test.go`
(`cd backend/services/traffic-police && go test ./...`).

## E2E scenario za proveru

1. Prijava policajca → `POST /api/auth/login`.
2. Registracija građanina, pa policajac evidentira vozača (`POST /api/drivers`
   sa `citizenId` iz registracije).
3. Policajac unosi prekršaj (`POST /api/violations`) → automatski nastaju
   kazna, poeni i obaveštenje.
4. Građanin vidi prekršaj i kaznu (`/api/me/...`), plati je
   (`POST /api/payments` + `/confirm`) → kazna plaćena, prekršaj rešen,
   stiže obaveštenje.
5. Unosom prekršaja dok poeni ne dostignu 18 → dozvola `suspended` + obaveštenje.

Isti tok je klikabilan kroz Angular aplikaciju (uloge građanin/policajac).
