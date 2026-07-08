package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"euprava/traffic-police/internal/repository"
	"euprava/traffic-police/internal/service"
)

// DriverHandler izlaže /drivers i /penalty-points rute.
type DriverHandler struct {
	drivers    *service.DriverService
	violations *service.ViolationService
}

func NewDriverHandler(drivers *service.DriverService, violations *service.ViolationService) *DriverHandler {
	return &DriverHandler{drivers: drivers, violations: violations}
}

func (h *DriverHandler) RegisterRoutes(r gin.IRouter) {
	g := r.Group("/drivers")
	g.GET("", h.list)
	g.POST("", h.create)
	g.GET("/:id", h.get)
	g.GET("/:id/violations", h.driverViolations)
	r.GET("/penalty-points/:driverId", h.penaltyPoints)
}

func (h *DriverHandler) list(c *gin.Context) {
	drivers, err := h.drivers.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, drivers)
}

func (h *DriverHandler) create(c *gin.Context) {
	var in service.RegisterDriverInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "neispravni podaci: " + err.Error()})
		return
	}
	d, err := h.drivers.Register(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, d)
}

func (h *DriverHandler) get(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	d, err := h.drivers.Get(c.Request.Context(), id)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *DriverHandler) driverViolations(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	violations, err := h.violations.ListByDriver(c.Request.Context(), id)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, violations)
}

func (h *DriverHandler) penaltyPoints(c *gin.Context) {
	id, ok := paramID(c, "driverId")
	if !ok {
		return
	}
	d, err := h.violations.PenaltyPoints(c.Request.Context(), id)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"driverId":      d.ID,
		"penaltyPoints": d.PenaltyPoints,
		"licenseStatus": d.LicenseStatus,
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

// respondErr mapira greške service sloja na HTTP odgovore; vraća true ako je greška obrađena.
func respondErr(c *gin.Context, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, repository.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "zapis nije pronađen"})
	case errors.Is(err, service.ErrUnknownViolationType), errors.Is(err, service.ErrFinePaid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	return true
}
