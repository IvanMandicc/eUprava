package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"euprava/vehicles/internal/repository"
	"euprava/vehicles/internal/service"
)

// VehicleHandler izlaže /vehicles rute i javnu proveru statusa po tablici.
type VehicleHandler struct {
	vehicles *service.VehicleService
}

func NewVehicleHandler(vehicles *service.VehicleService) *VehicleHandler {
	return &VehicleHandler{vehicles: vehicles}
}

func (h *VehicleHandler) RegisterRoutes(r gin.IRouter) {
	g := r.Group("/vehicles")
	g.GET("", h.list)
	g.POST("", h.register)
	g.GET("/:id", h.get)
	g.POST("/:id/transfer", h.transfer)
	g.POST("/:id/renew-registration", h.renew)
	g.POST("/:id/report-theft", h.reportTheft)
	g.POST("/:id/report-found", h.reportFound)
	r.GET("/vehicle-status/:plate", h.plateStatus)
	// owner-by-plate je za razliku od vehicle-status namenjen policajcu (ne
	// javnosti) — koristi ga Traffic Police servis kad prekršaj snimi kamera
	// i zna se samo registarska tablica, ne i vozač.
	g.GET("/owner-by-plate/:plate", h.ownerByPlate)
}

func (h *VehicleHandler) list(c *gin.Context) {
	vehicles, err := h.vehicles.List(c.Request.Context())
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, vehicles)
}

func (h *VehicleHandler) register(c *gin.Context) {
	var in service.RegisterVehicleInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "neispravni podaci: " + err.Error()})
		return
	}
	v, err := h.vehicles.Register(c.Request.Context(), in)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusCreated, v)
}

func (h *VehicleHandler) get(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	v, err := h.vehicles.Get(c.Request.Context(), id)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *VehicleHandler) transfer(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	var in service.TransferVehicleInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "neispravni podaci: " + err.Error()})
		return
	}
	v, err := h.vehicles.Transfer(c.Request.Context(), id, in)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *VehicleHandler) renew(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	v, err := h.vehicles.RenewRegistration(c.Request.Context(), id)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *VehicleHandler) reportTheft(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	cid, role, ok := requester(c)
	if !ok {
		return
	}
	report, err := h.vehicles.ReportTheft(c.Request.Context(), id, cid, role)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusCreated, report)
}

func (h *VehicleHandler) reportFound(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	cid, role, ok := requester(c)
	if !ok {
		return
	}
	v, err := h.vehicles.ReportFound(c.Request.Context(), id, cid, role)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, v)
}

// plateStatus je brza provera statusa vozila po tablici (npr. za buduću
// integraciju sa pretragama Traffic Police servisa pri saobraćajnoj kontroli).
func (h *VehicleHandler) plateStatus(c *gin.Context) {
	plate := c.Param("plate")
	v, err := h.vehicles.GetByPlate(c.Request.Context(), plate)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"plateNumber":            v.PlateNumber,
		"status":                 v.Status,
		"registrationValidUntil": v.RegistrationValidUntil,
		"make":                   v.Make,
		"model":                  v.Model,
	})
}

// ownerByPlate vraća vlasnika vozila po tablici — za razliku od plateStatus
// (javna, anonimna provera) ovo je interna/policijska ruta koja otkriva
// identitet vlasnika (citizenId), pa je dostupna samo službeniku.
func (h *VehicleHandler) ownerByPlate(c *gin.Context) {
	plate := c.Param("plate")
	v, err := h.vehicles.GetByPlate(c.Request.Context(), plate)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ownerCitizenId": v.OwnerCitizenID,
		"plateNumber":    v.PlateNumber,
		"vin":            v.VIN,
		"make":           v.Make,
		"model":          v.Model,
	})
}

// ---------------- zajedničke pomoćne funkcije ----------------

func paramID(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "neispravan " + name})
		return 0, false
	}
	return id, true
}

// requester čita identitet pozivaoca iz header-a koje postavlja API Gateway.
func requester(c *gin.Context) (int64, string, bool) {
	id, err := strconv.ParseInt(c.GetHeader("X-User-Id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "nedostaje X-User-Id header"})
		return 0, "", false
	}
	return id, c.GetHeader("X-User-Role"), true
}

// respondErr mapira greške service sloja na HTTP odgovore; vraća true ako je greška obrađena.
func respondErr(c *gin.Context, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, repository.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "zapis nije pronađen"})
	case errors.Is(err, service.ErrOwnerNotCitizen),
		errors.Is(err, service.ErrDuplicateVIN),
		errors.Is(err, service.ErrInvalidVIN),
		errors.Is(err, service.ErrSameOwner),
		errors.Is(err, service.ErrTechInspectionExpired),
		errors.Is(err, service.ErrInsuranceExpired),
		errors.Is(err, service.ErrNotStolen),
		errors.Is(err, service.ErrInvalidPlateFormat),
		errors.Is(err, service.ErrPlateForbidden),
		errors.Is(err, service.ErrPlateTaken),
		errors.Is(err, service.ErrReservationNotPending):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrVehicleStolen), errors.Is(err, service.ErrBlockedByFines):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrNotOwner):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	return true
}
