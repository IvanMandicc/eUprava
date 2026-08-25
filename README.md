# eUprava (mikroservisna arhitektura)

Timski studentski projekat po uzoru na portal eUprava, izrađen za predmet
Tehnologije i sistemi eUprave. Dve teme, po jedna po članu tima —
**Saobraćajna policija** i **MUP — Vozila** — deljene su preko zajedničkog
sistema korisnika, SSO autentifikacije (JWT) i API gateway-a. Sistem čini
šest servisa koji komuniciraju preko REST API-ja, svaki sa sopstvenom
PostgreSQL bazom.

## Arhitektura

```
              Angular (:4200)              React — Vozila (:4201)
                     |                              |
                     +--------------+---------------+
                                    |
                             API Gateway (:8080)
                                    |
     +------------+-----------+-----------+-------------+-------------+
     |            |                       |             |             |
 Citizen      Traffic Police          Payment      Notification    Vehicles
 (:8081)        (:8082)               (:8083)        (:8084)        (:8085)
     |            |                       |             |             |
 citizen_db   traffic_db             payment_db     notification_db vehicle_db
```

| Servis | Port | Odgovornost |
|---|---|---|
| API Gateway | 8080 | Jedina ulazna tačka: JWT validacija, autorizacija po ulozi, reverse proxy, CORS |
| Citizen | 8081 | Registracija/prijava (JWT) — deljeni sistem korisnika i SSO za ceo tim |
| **Traffic Police** | 8082 | Vozači, prekršaji, novčane kazne, kazneni poeni, status dozvole |
| Payment | 8083 | Iniciranje i potvrda plaćanja kazni |
| Notification | 8084 | Elektronska obaveštenja građanima |
| **Vehicles (MUP-vozila)** | 8085 | Registar vozila, prenos vlasništva, produženje registracije, krađa, personalizovane tablice, digitalni izveštaji |

Komunikacija servis–servis (interna, van gateway-a):
- Traffic Police → Citizen: provera/dobavljanje podataka o građaninu
- Traffic Police → Notification: obaveštenje o novoj kazni / suspenziji dozvole
- Payment → Traffic Police: `GET /fines/{id}` (provera) i `PUT /fines/{id}/pay` (evidencija plaćanja)
- Payment → Notification: potvrda uplate
- Vehicles → Citizen: `GET /citizens/{id}` — verifikacija vlasnika pri registraciji/prenosu vozila
- Vehicles → Traffic Police: `GET /fines/citizen/{id}/unpaid-summary` — provera neplaćenih kazni pre prenosa vlasništva ili produženja registracije
- Vehicles → Notification: `POST /notifications` — obaveštenja o registraciji, prenosu, krađi i tablicama

Deljeni `Citizen` servis i JWT (izdat pri prijavi, validiran na gateway-u za
sve servise) predstavljaju tim-nivo sistem korisnika i SSO metodu
autentifikacije — oba frontenda (Angular i React) pozivaju isti
`POST /api/auth/login` i koriste token istog formata.

## Tehnologije

Go (Gin) · Angular 19 i React 18 (Vite) · PostgreSQL 16 (baza po servisu) · Docker Compose · JWT (HS256)

## Struktura repozitorijuma

```
eUprava/
├── docker-compose.yml
├── backend/
│   ├── api-gateway/            # ulazna tačka, JWT + role middleware, proxy
│   └── services/
│       ├── citizen/            # auth i podaci o građanima (deljeni sistem korisnika/SSO)
│       ├── traffic-police/     # tema 1 — Saobraćajna policija
│       ├── payment/            # plaćanja kazni
│       ├── notification/       # obaveštenja
│       └── vehicles/           # tema 2 — MUP-vozila
│           (svaki servis: cmd/main.go + internal/{handler,service,repository,model,client})
├── frontend/                   # Angular aplikacija — Saobraćajna policija
└── frontend-vehicles/          # React aplikacija — MUP-vozila
```

## Pokretanje

```bash
cp .env.example .env    
docker compose up --build
```

- Frontend (Saobraćajna policija): http://localhost:4200
- Frontend (MUP-vozila): http://localhost:4201
- API Gateway: http://localhost:8080

**Podrazumevani nalozi** (seed-uju se automatski pri startu):
- policajac: `policija@euprava.rs` / `policija123`
- administrator: `admin@euprava.rs` / `admin123`

Građani se registruju kroz aplikaciju (uloga `citizen`); naloge policajaca
kreira administrator kroz stranicu "Korisnici". Isti nalog policajca
(`policija@euprava.rs`) radi i u Traffic Police delu sistema i u MUP-vozila
delu — obe teme dele istu ulogu `officer` (jedan sistem korisnika, SSO).

### Pokretanje bez Dockera (razvoj)

Svaki servis je samostalan Go modul — potrebne su mu env promenljive
`DATABASE_URL`, `PORT` i URL-ovi servisa sa kojima komunicira (podrazumevane
vrednosti su u `cmd/main.go`). Frontend (Traffic Police): `cd frontend && npm install && npm start`.
Frontend (Vozila): `cd frontend-vehicles && npm install && npm run dev`.

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

## Poslovna pravila (Vehicles — MUP-vozila)

1. **Prva registracija vozila** (`POST /vehicles`) — proverava vlasnika kod
   Citizen servisa (mora imati ulogu `citizen`), validira jedinstven VIN i
   **automatski generiše registarsku tablicu** po šablonu `XX-NNN-XX`
   (gradska oznaka + 3 cifre + 2 slova), uz proveru kolizije.
2. **Prenos vlasništva** (`POST /vehicles/{id}/transfer`) — blokira se ako je
   vozilo prijavljeno kao ukradeno ili ako trenutni vlasnik ima **2+
   neplaćene kazne ili 30.000+ RSD duga** (provera kod Traffic Police
   servisa); nova registracija postaje obavezna kod novog vlasnika.
3. **Produženje registracije** (`POST /vehicles/{id}/renew-registration`) —
   odbija se ako je istekao tehnički pregled ili osiguranje, ili ako postoje
   iste blokirajuće neplaćene kazne; produžava rok za godinu dana.
4. **Prijava/pronalazak krađe** (`POST /vehicles/{id}/report-theft`,
   `POST /vehicles/{id}/report-found`) — samo vlasnik ili službenik; status
   `stolen` blokira transfer i produženje registracije.
5. **Personalizovane tablice** (`POST /plate-reservations`) — validacija
   formata (4-7 znakova) i zabranjenih reči, provera kolizije sa postojećim
   tablicama/rezervacijama, viša taksa za kombinacije do 5 znakova,
   odobravanje/odbijanje od strane službenika.
6. **Digitalni izveštaj o vozilu** (`POST /vehicles/{id}/reports`) — generiše
   istorijat vlasništva sa nasumičnim verifikacionim kodom, proverljivim
   javno bez prijave (`GET /reports/verify/{code}`).

Unit testovi poslovnih pravila: `backend/services/vehicles/internal/service/`
(`cd backend/services/vehicles && go test ./...`).

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
| `GET/POST /api/vehicles`, `GET /api/vehicles/{id}` | policajac | registar vozila |
| `POST /api/vehicles/{id}/transfer`, `/renew-registration` | policajac | prenos vlasništva, produženje registracije |
| `POST /api/vehicles/{id}/report-theft`, `/report-found`, `/reports` | građanin (vlasnik) ili policajac | prijava krađe/pronalaska, generisanje izveštaja |
| `GET /api/vehicle-status/{plate}` | policajac | brza provera statusa vozila po tablici |
| `GET /api/me/vehicles` | građanin | sopstvena vozila |
| `POST /api/plate-reservations` | građanin | zahtev za personalizovanu tablicu |
| `GET /api/plate-reservations`, `PUT /api/plate-reservations/{id}/decision` | policajac | red čekanja i odluka o zahtevu |
| `GET /api/me/plate-reservations` | građanin | sopstveni zahtevi za tablice |
| `GET /api/reports/verify/{code}` | **javno** | verifikacija autentičnosti izveštaja o vozilu |

Interni endpointi (nisu dostupni spolja): `PUT /fines/{id}/pay` (Payment → Traffic
Police), `POST /notifications` (servisi → Notification), `GET /fines/citizen/{id}/unpaid-summary`
(Vehicles → Traffic Police).

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

## E2E scenario za proveru (Vehicles)

1. Prijava policajca → `POST /api/auth/login` (isti nalog kao za Traffic Police).
2. Registracija građanina (ako već ne postoji), pa policajac registruje
   vozilo na njegovo ime (`POST /api/vehicles`) → tablica se automatski
   generiše.
3. Građanin vidi vozilo (`GET /api/me/vehicles`), zatraži personalizovanu
   tablicu (`POST /api/plate-reservations`) → policajac je odobrava
   (`PUT /api/plate-reservations/{id}/decision`) → građanin dobija obaveštenje.
4. Građanin prijavi krađu vozila (`POST /api/vehicles/{id}/report-theft`) →
   status postaje `stolen`; pokušaj prenosa vlasništva u tom stanju vraća grešku.
5. Nakon pronalaska (`POST /api/vehicles/{id}/report-found`), policajac
   produžava registraciju (`POST /api/vehicles/{id}/renew-registration`).
6. Građanin generiše digitalni izveštaj o vozilu (`POST /api/vehicles/{id}/reports`)
   i bilo ko može proveriti kod bez prijave (`GET /api/reports/verify/{code}`).

Isti tok je klikabilan kroz React aplikaciju na portu 4201 (uloge građanin/policajac).
