package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"euprava/vehicles/internal/client"
	"euprava/vehicles/internal/model"
	"euprava/vehicles/internal/repository"
)

var (
	ErrOwnerNotCitizen       = errors.New("vlasnik mora biti građanin (uloga citizen)")
	ErrDuplicateVIN          = errors.New("vozilo sa ovim brojem šasije već postoji")
	ErrInvalidVIN            = errors.New("broj šasije (VIN) mora imati 17 alfanumeričkih znakova")
	ErrVehicleStolen         = errors.New("vozilo je prijavljeno kao ukradeno")
	ErrNotStolen             = errors.New("vozilo nije prijavljeno kao ukradeno")
	ErrBlockedByFines        = errors.New("nije moguće izvršiti radnju: vlasnik ima neplaćene kazne iznad dozvoljenog praga")
	ErrTechInspectionExpired = errors.New("tehnički pregled je istekao — obnovite pre produženja registracije")
	ErrInsuranceExpired      = errors.New("polisa osiguranja je istekla — obnovite pre produženja registracije")
	ErrSameOwner             = errors.New("novi vlasnik je isti kao trenutni vlasnik")
	ErrNotOwner              = errors.New("samo vlasnik vozila ili službenik MUP-a mogu izvršiti ovu radnju")
)

var vinPattern = regexp.MustCompile(`^[A-HJ-NPR-Z0-9]{17}$`)

// plateCities je fond gradskih oznaka korišćen pri automatskom generisanju
// registarske tablice; plateLetters isključuje slova kojih nema u srpskim
// tablicama (Q, W, X, Y).
var plateCities = []string{"BG", "NS", "NI", "KG", "SU", "CA", "PA", "LE", "VA", "ZR"}

const plateLetters = "ABCDEFGHIJKLMNOPRSTUVZ"

// RegisterVehicleInput su podaci za prvu registraciju vozila.
type RegisterVehicleInput struct {
	OwnerCitizenID           int64     `json:"ownerCitizenId" binding:"required"`
	VIN                      string    `json:"vin" binding:"required"`
	Make                     string    `json:"make" binding:"required"`
	Model                    string    `json:"model" binding:"required"`
	Year                     int       `json:"year" binding:"required,min=1900"`
	Category                 string    `json:"category" binding:"required,oneof=M1 M2 M3 N1 N2 N3 L"`
	Color                    string    `json:"color"`
	EnginePowerKw            int       `json:"enginePowerKw" binding:"min=0"`
	FuelType                 string    `json:"fuelType" binding:"required,oneof=benzin dizel hibrid elektro gas"`
	InsuranceValidUntil      time.Time `json:"insuranceValidUntil" binding:"required"`
	TechInspectionValidUntil time.Time `json:"techInspectionValidUntil" binding:"required"`
}

// TransferVehicleInput su podaci za prenos vlasništva.
type TransferVehicleInput struct {
	NewOwnerCitizenID int64 `json:"newOwnerCitizenId" binding:"required"`
}

// VehicleService sadrži centralnu poslovnu logiku registra vozila:
// registracija, prenos vlasništva, produženje registracije i status krađe.
type VehicleService struct {
	vehicles  repository.VehicleRepository
	transfers repository.TransferRepository
	thefts    repository.TheftRepository
	citizens  client.CitizenClient
	traffic   client.TrafficClient
	notifier  client.NotificationClient
}

func NewVehicleService(
	vehicles repository.VehicleRepository,
	transfers repository.TransferRepository,
	thefts repository.TheftRepository,
	citizens client.CitizenClient,
	traffic client.TrafficClient,
	notifier client.NotificationClient,
) *VehicleService {
	return &VehicleService{
		vehicles: vehicles, transfers: transfers, thefts: thefts,
		citizens: citizens, traffic: traffic, notifier: notifier,
	}
}

// Register evidentira novo vozilo: proverava vlasnika kod Citizen servisa,
// validira jedinstven VIN i automatski generiše registarsku tablicu.
func (s *VehicleService) Register(ctx context.Context, in RegisterVehicleInput) (*model.Vehicle, error) {
	vin := strings.ToUpper(strings.TrimSpace(in.VIN))
	if !vinPattern.MatchString(vin) {
		return nil, ErrInvalidVIN
	}

	owner, err := s.citizens.GetCitizen(ctx, in.OwnerCitizenID)
	if err != nil {
		return nil, err
	}
	if owner.Role != "citizen" {
		return nil, ErrOwnerNotCitizen
	}

	if _, err := s.vehicles.GetByVIN(ctx, vin); err == nil {
		return nil, ErrDuplicateVIN
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	plate, err := s.generatePlate(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	v := &model.Vehicle{
		OwnerCitizenID:           in.OwnerCitizenID,
		VIN:                      vin,
		PlateNumber:              plate,
		Make:                     in.Make,
		Model:                    in.Model,
		Year:                     in.Year,
		Category:                 in.Category,
		Color:                    in.Color,
		EnginePowerKw:            in.EnginePowerKw,
		FuelType:                 in.FuelType,
		FirstRegistrationDate:    now,
		InsuranceValidUntil:      in.InsuranceValidUntil,
		TechInspectionValidUntil: in.TechInspectionValidUntil,
		RegistrationValidUntil:   now.Add(model.RegistrationValidityPeriod),
		Status:                   model.VehicleRegistered,
	}
	if err := s.vehicles.Create(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *VehicleService) generatePlate(ctx context.Context) (string, error) {
	for attempt := 0; attempt < 25; attempt++ {
		city := plateCities[rand.Intn(len(plateCities))]
		digits := rand.Intn(1000)
		letters := randomLetters(2)
		plate := fmt.Sprintf("%s-%03d-%s", city, digits, letters)

		exists, err := s.vehicles.PlateExists(ctx, plate)
		if err != nil {
			return "", err
		}
		if !exists {
			return plate, nil
		}
	}
	return "", errors.New("nije moguće generisati jedinstvenu tablicu, pokušajte ponovo")
}

func randomLetters(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = plateLetters[rand.Intn(len(plateLetters))]
	}
	return string(b)
}

// Get vraća vozilo obogaćeno podacima o vlasniku iz Citizen servisa.
func (s *VehicleService) Get(ctx context.Context, id int64) (*model.VehicleDetails, error) {
	v, err := s.vehicles.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	details := &model.VehicleDetails{Vehicle: *v}
	// Ako Citizen servis trenutno nije dostupan, vraćamo vozilo bez detalja
	// o vlasniku umesto greške (labava povezanost servisa, isti pristup kao
	// kod Traffic Police → Citizen pozivа).
	if info, err := s.citizens.GetCitizen(ctx, v.OwnerCitizenID); err == nil {
		details.Owner = info
	} else {
		log.Printf("citizen servis nedostupan za citizenId=%d: %v", v.OwnerCitizenID, err)
	}
	return details, nil
}

// GetByPlate vraća vozilo po registarskoj tablici (brza provera statusa).
func (s *VehicleService) GetByPlate(ctx context.Context, plate string) (*model.Vehicle, error) {
	return s.vehicles.GetByPlate(ctx, strings.ToUpper(strings.TrimSpace(plate)))
}

// ListByOwner vraća vozila u vlasništvu građanina (za /me rute).
func (s *VehicleService) ListByOwner(ctx context.Context, citizenID int64) ([]model.Vehicle, error) {
	return s.vehicles.ListByOwner(ctx, citizenID)
}

// List vraća sva vozila (za službenika).
func (s *VehicleService) List(ctx context.Context) ([]model.Vehicle, error) {
	return s.vehicles.List(ctx)
}

// Transfer prenosi vlasništvo nad vozilom na novog građanina: verifikuje
// novog vlasnika, blokira ukradena vozila i vlasnike sa neplaćenim kaznama
// iznad praga (provera kod Traffic Police servisa), i nalaže ponovnu
// registraciju kod novog vlasnika.
func (s *VehicleService) Transfer(ctx context.Context, vehicleID int64, in TransferVehicleInput) (*model.Vehicle, error) {
	v, err := s.vehicles.GetByID(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	if v.Status == model.VehicleStolen {
		return nil, ErrVehicleStolen
	}
	if in.NewOwnerCitizenID == v.OwnerCitizenID {
		return nil, ErrSameOwner
	}

	newOwner, err := s.citizens.GetCitizen(ctx, in.NewOwnerCitizenID)
	if err != nil {
		return nil, err
	}
	if newOwner.Role != "citizen" {
		return nil, ErrOwnerNotCitizen
	}

	if err := s.checkUnpaidFines(ctx, v.OwnerCitizenID); err != nil {
		return nil, err
	}

	oldOwnerID := v.OwnerCitizenID
	transfer := &model.OwnershipTransfer{
		VehicleID: v.ID, FromCitizenID: oldOwnerID, ToCitizenID: in.NewOwnerCitizenID, TransferDate: time.Now(),
	}
	if err := s.transfers.Create(ctx, transfer); err != nil {
		return nil, err
	}

	v.OwnerCitizenID = in.NewOwnerCitizenID
	// Nova registracija je obavezna nakon prenosa vlasništva — rok se poništava.
	v.RegistrationValidUntil = time.Now()
	if err := s.vehicles.Update(ctx, v); err != nil {
		return nil, err
	}

	s.notify(ctx, oldOwnerID, fmt.Sprintf("Vozilo %s je prepisano na novog vlasnika.", v.PlateNumber), "VEHICLE_TRANSFERRED")
	s.notify(ctx, in.NewOwnerCitizenID,
		fmt.Sprintf("Vozilo %s je prepisano na vas. Neophodno je da izvršite novu registraciju.", v.PlateNumber),
		"VEHICLE_TRANSFERRED")
	return v, nil
}

// RenewRegistration produžava registraciju vozila za godinu dana — odbija
// ako je istekao tehnički pregled/osiguranje ili ako postoje blokirajuće
// neplaćene kazne kod vlasnika.
func (s *VehicleService) RenewRegistration(ctx context.Context, vehicleID int64) (*model.Vehicle, error) {
	v, err := s.vehicles.GetByID(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	if v.Status == model.VehicleStolen {
		return nil, ErrVehicleStolen
	}

	now := time.Now()
	if v.TechInspectionValidUntil.Before(now) {
		return nil, ErrTechInspectionExpired
	}
	if v.InsuranceValidUntil.Before(now) {
		return nil, ErrInsuranceExpired
	}
	if err := s.checkUnpaidFines(ctx, v.OwnerCitizenID); err != nil {
		return nil, err
	}

	v.RegistrationValidUntil = now.Add(model.RegistrationValidityPeriod)
	if err := s.vehicles.Update(ctx, v); err != nil {
		return nil, err
	}

	s.notify(ctx, v.OwnerCitizenID,
		fmt.Sprintf("Registracija vozila %s je produžena do %s.", v.PlateNumber, v.RegistrationValidUntil.Format("02.01.2006.")),
		"REGISTRATION_RENEWED")
	return v, nil
}

// checkUnpaidFines proverava kod Traffic Police servisa da li vlasnik ima
// neplaćene kazne iznad praga. Ako servis nije dostupan, radnja se ne
// blokira (labava povezanost) — greška se samo loguje.
func (s *VehicleService) checkUnpaidFines(ctx context.Context, citizenID int64) error {
	summary, err := s.traffic.UnpaidFinesSummary(ctx, citizenID)
	if err != nil {
		log.Printf("traffic police servis nedostupan za citizenId=%d: %v", citizenID, err)
		return nil
	}
	if summary.UnpaidCount >= model.UnpaidFinesCountThreshold || summary.UnpaidTotal >= model.UnpaidFinesTotalThreshold {
		return ErrBlockedByFines
	}
	return nil
}

// ReportTheft prijavljuje vozilo kao ukradeno — samo vlasnik ili službenik
// mogu prijaviti; blokira dalje transfere i produženja registracije.
func (s *VehicleService) ReportTheft(ctx context.Context, vehicleID, requesterCitizenID int64, requesterRole string) (*model.TheftReport, error) {
	v, err := s.vehicles.GetByID(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	if requesterRole == "citizen" && requesterCitizenID != v.OwnerCitizenID {
		return nil, ErrNotOwner
	}
	if v.Status == model.VehicleStolen {
		return nil, ErrVehicleStolen
	}

	report := &model.TheftReport{
		VehicleID: vehicleID, ReportedByID: requesterCitizenID, Status: model.TheftReported, ReportedAt: time.Now(),
	}
	if err := s.thefts.Create(ctx, report); err != nil {
		return nil, err
	}

	v.Status = model.VehicleStolen
	if err := s.vehicles.Update(ctx, v); err != nil {
		return nil, err
	}

	s.notify(ctx, v.OwnerCitizenID, fmt.Sprintf("Vozilo %s je evidentirano kao ukradeno.", v.PlateNumber), "VEHICLE_STOLEN")
	return report, nil
}

// ReportFound vraća status vozila na registrovano nakon pronalaska.
func (s *VehicleService) ReportFound(ctx context.Context, vehicleID, requesterCitizenID int64, requesterRole string) (*model.Vehicle, error) {
	v, err := s.vehicles.GetByID(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	if requesterRole == "citizen" && requesterCitizenID != v.OwnerCitizenID {
		return nil, ErrNotOwner
	}
	if v.Status != model.VehicleStolen {
		return nil, ErrNotStolen
	}

	report, err := s.thefts.GetOpenByVehicle(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	if err := s.thefts.MarkFound(ctx, report.ID); err != nil {
		return nil, err
	}

	v.Status = model.VehicleRegistered
	if err := s.vehicles.Update(ctx, v); err != nil {
		return nil, err
	}

	s.notify(ctx, v.OwnerCitizenID, fmt.Sprintf("Vozilo %s je pronađeno — status je vraćen na registrovano.", v.PlateNumber), "VEHICLE_FOUND")
	return v, nil
}

func (s *VehicleService) notify(ctx context.Context, citizenID int64, message, notifType string) {
	if err := s.notifier.Notify(ctx, citizenID, message, notifType); err != nil {
		log.Printf("slanje obaveštenja nije uspelo (citizenId=%d): %v", citizenID, err)
	}
}
