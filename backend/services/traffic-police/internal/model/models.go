package model

import "time"

// LicenseStatus je status vozačke dozvole.
type LicenseStatus string

const (
	LicenseValid     LicenseStatus = "valid"
	LicenseSuspended LicenseStatus = "suspended"
)

// PenaltyPointsLimit — kada vozač dostigne ovaj broj poena, dozvola se suspenduje.
const PenaltyPointsLimit = 18

// ViolationStatus je status prekršaja.
type ViolationStatus string

const (
	ViolationActive   ViolationStatus = "active"
	ViolationResolved ViolationStatus = "resolved"
)

// Driver je evidencija vozača — čuva samo referencu na građanina (citizenId),
// detaljne podatke o građaninu daje Citizen servis.
type Driver struct {
	ID            int64         `json:"id"`
	CitizenID     int64         `json:"citizenId"`
	LicenseNumber string        `json:"licenseNumber"`
	PenaltyPoints int           `json:"penaltyPoints"`
	LicenseStatus LicenseStatus `json:"licenseStatus"`
}

// CitizenInfo su podaci o građaninu dobijeni od Citizen servisa.
type CitizenInfo struct {
	ID        int64  `json:"id"`
	JMBG      string `json:"jmbg"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Address   string `json:"address"`
}

// DriverDetails je vozač obogaćen podacima o građaninu.
type DriverDetails struct {
	Driver
	Citizen *CitizenInfo `json:"citizen,omitempty"`
}

// Violation je saobraćajni prekršaj.
type Violation struct {
	ID          int64           `json:"id"`
	DriverID    int64           `json:"driverId"`
	Type        string          `json:"type"`
	Description string          `json:"description"`
	Date        time.Time       `json:"date"`
	Location    string          `json:"location"`
	Points      int             `json:"points"`
	FineAmount  float64         `json:"fineAmount"`
	Status      ViolationStatus `json:"status"`
}

// Fine je novčana kazna vezana za prekršaj.
type Fine struct {
	ID              int64      `json:"id"`
	ViolationID     int64      `json:"violationId"`
	Amount          float64    `json:"amount"`
	PaymentDeadline time.Time  `json:"paymentDeadline"`
	Paid            bool       `json:"paid"`
	PaymentDate     *time.Time `json:"paymentDate"`
}

// FineDetails je kazna obogaćena podacima o prekršaju i vozaču
// (citizenId koristi Payment servis da zna ko plaća).
type FineDetails struct {
	Fine
	ViolationType string    `json:"violationType"`
	ViolationDate time.Time `json:"violationDate"`
	Location      string    `json:"location"`
	DriverID      int64     `json:"driverId"`
	CitizenID     int64     `json:"citizenId"`
}

// ViolationStat je agregirana, anonimna statistika prekršaja po tipu —
// javno dostupan podatak (open data), bez ijednog ličnog podatka.
type ViolationStat struct {
	Type        string  `json:"type"`
	Count       int64   `json:"count"`
	TotalPoints int64   `json:"totalPoints"`
	AvgFine     float64 `json:"avgFine"`
	TotalFines  float64 `json:"totalFines"`
}

// ViolationTypeInfo opisuje tip prekršaja sa podrazumevanim poenima i iznosom kazne.
type ViolationTypeInfo struct {
	Code   string  `json:"code"`
	Label  string  `json:"label"`
	Points int     `json:"points"`
	Fine   float64 `json:"fine"`
}

// ViolationCatalog je šifarnik tipova prekršaja.
var ViolationCatalog = []ViolationTypeInfo{
	{Code: "SPEEDING", Label: "Prekoračenje brzine", Points: 6, Fine: 20000},
	{Code: "RED_LIGHT", Label: "Prolazak kroz crveno svetlo", Points: 8, Fine: 25000},
	{Code: "NO_SEATBELT", Label: "Nevezan pojas", Points: 2, Fine: 5000},
	{Code: "DRUNK_DRIVING", Label: "Vožnja pod dejstvom alkohola", Points: 14, Fine: 100000},
	{Code: "PHONE_USAGE", Label: "Upotreba telefona u vožnji", Points: 3, Fine: 10000},
	{Code: "ILLEGAL_PARKING", Label: "Nepropisno parkiranje", Points: 1, Fine: 3000},
}

// CatalogEntry vraća tip prekršaja iz šifarnika po šifri.
func CatalogEntry(code string) (ViolationTypeInfo, bool) {
	for _, v := range ViolationCatalog {
		if v.Code == code {
			return v, true
		}
	}
	return ViolationTypeInfo{}, false
}
