package model

import "time"

// VehicleStatus je status vozila u registru.
type VehicleStatus string

const (
	VehicleRegistered VehicleStatus = "registered"
	VehicleStolen     VehicleStatus = "stolen"
)

// Poslovne konstante — pragovi i rokovi (isti duh kao PenaltyPointsLimit
// kod Traffic Police servisa).
const (
	// RegistrationValidityPeriod je trajanje registracije vozila (1 godina).
	RegistrationValidityPeriod = 365 * 24 * time.Hour

	// UnpaidFinesCountThreshold i UnpaidFinesTotalThreshold — kada vlasnik
	// ima ovoliko ili više neplaćenih kazni (Traffic Police), transfer i
	// produženje registracije se blokiraju.
	UnpaidFinesCountThreshold = 2
	UnpaidFinesTotalThreshold = 30000.0

	// VanityPlateReservationTTL — nepotvrđena rezervacija tablice ističe
	// nakon ovog perioda.
	VanityPlateReservationTTL = 3 * 24 * time.Hour

	// VanityPlateBaseFee je osnovna taksa za personalizovanu tablicu;
	// kraće/traženije kombinacije (do 5 znakova) plaćaju višu taksu.
	VanityPlateBaseFee     = 5000.0
	VanityPlateShortFee    = 15000.0
	VanityPlateShortMaxLen = 5
)

// Vehicle je vozilo u registru MUP-a.
type Vehicle struct {
	ID                       int64         `json:"id"`
	OwnerCitizenID           int64         `json:"ownerCitizenId"`
	VIN                      string        `json:"vin"`
	PlateNumber              string        `json:"plateNumber"`
	Make                     string        `json:"make"`
	Model                    string        `json:"model"`
	Year                     int           `json:"year"`
	Category                 string        `json:"category"`
	Color                    string        `json:"color"`
	EnginePowerKw            int           `json:"enginePowerKw"`
	FuelType                 string        `json:"fuelType"`
	FirstRegistrationDate    time.Time     `json:"firstRegistrationDate"`
	InsuranceValidUntil      time.Time     `json:"insuranceValidUntil"`
	TechInspectionValidUntil time.Time     `json:"techInspectionValidUntil"`
	RegistrationValidUntil   time.Time     `json:"registrationValidUntil"`
	Status                   VehicleStatus `json:"status"`
	CreatedAt                time.Time     `json:"createdAt"`
}

// CitizenInfo su podaci o vlasniku dobijeni od Citizen servisa.
type CitizenInfo struct {
	ID        int64  `json:"id"`
	JMBG      string `json:"jmbg"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Address   string `json:"address"`
}

// VehicleDetails je vozilo obogaćeno podacima o vlasniku.
type VehicleDetails struct {
	Vehicle
	Owner *CitizenInfo `json:"owner,omitempty"`
}

// UnpaidFinesSummary je odgovor Traffic Police servisa na proveru
// neplaćenih kazni vlasnika (interna razmena podataka).
type UnpaidFinesSummary struct {
	UnpaidCount int     `json:"unpaidCount"`
	UnpaidTotal float64 `json:"unpaidTotal"`
}

// OwnershipTransfer je zapis o prenosu vlasništva nad vozilom.
type OwnershipTransfer struct {
	ID            int64     `json:"id"`
	VehicleID     int64     `json:"vehicleId"`
	FromCitizenID int64     `json:"fromCitizenId"`
	ToCitizenID   int64     `json:"toCitizenId"`
	TransferDate  time.Time `json:"transferDate"`
}

// TheftReportStatus je status prijave krađe.
type TheftReportStatus string

const (
	TheftReported TheftReportStatus = "reported"
	TheftFound    TheftReportStatus = "found"
)

// TheftReport je prijava krađe/pronalaska vozila.
type TheftReport struct {
	ID           int64             `json:"id"`
	VehicleID    int64             `json:"vehicleId"`
	ReportedByID int64             `json:"reportedByCitizenId"`
	Status       TheftReportStatus `json:"status"`
	ReportedAt   time.Time         `json:"reportedAt"`
	FoundAt      *time.Time        `json:"foundAt"`
}

// PlateReservationStatus je status zahteva za personalizovanu tablicu.
type PlateReservationStatus string

const (
	PlateReservationPending  PlateReservationStatus = "pending"
	PlateReservationApproved PlateReservationStatus = "approved"
	PlateReservationRejected PlateReservationStatus = "rejected"
	PlateReservationExpired  PlateReservationStatus = "expired"
)

// PlateReservation je zahtev građanina za personalizovanu registarsku tablicu.
type PlateReservation struct {
	ID             int64                  `json:"id"`
	RequestedByID  int64                  `json:"requestedByCitizenId"`
	RequestedPlate string                 `json:"requestedPlate"`
	FeeAmount      float64                `json:"feeAmount"`
	Status         PlateReservationStatus `json:"status"`
	RequestedAt    time.Time              `json:"requestedAt"`
	DecidedAt      *time.Time             `json:"decidedAt"`
}

// VehicleReport je generisani digitalni izveštaj o vozilu, proverljiv
// preko javnog verifikacionog koda.
type VehicleReport struct {
	ID               int64     `json:"id"`
	VehicleID        int64     `json:"vehicleId"`
	VerificationCode string    `json:"verificationCode"`
	GeneratedAt      time.Time `json:"generatedAt"`
}

// VehicleReportDetails je pun sadržaj izveštaja — vozilo i istorijat
// prenosa vlasništva.
type VehicleReportDetails struct {
	VehicleReport
	Vehicle   Vehicle             `json:"vehicle"`
	Transfers []OwnershipTransfer `json:"transfers"`
}
