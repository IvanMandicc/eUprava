package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"euprava/traffic-police/internal/repository"
	"euprava/traffic-police/internal/service"
)

// MeHandler izlaže /me rute — građanin pregleda sopstvene podatke.
// Identitet stiže kroz X-User-Id header koji postavlja API Gateway.
type MeHandler struct {
	drivers    *service.DriverService
	violations *service.ViolationService
	fines      *service.FineService
}

func NewMeHandler(drivers *service.DriverService, violations *service.ViolationService, fines *service.FineService) *MeHandler {
	return &MeHandler{drivers: drivers, violations: violations, fines: fines}
}

func (h *MeHandler) RegisterRoutes(r gin.IRouter) {
	g := r.Group("/me")
	g.GET("/driver", h.driver)
	g.GET("/violations", h.violationList)
	g.GET("/fines", h.fineList)
}

func citizenID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.GetHeader("X-User-Id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "nedostaje X-User-Id header"})
		return 0, false
	}
	return id, true
}

func (h *MeHandler) driver(c *gin.Context) {
	cid, ok := citizenID(c)
	if !ok {
		return
	}
	d, err := h.drivers.GetByCitizen(c.Request.Context(), cid)
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "niste evidentirani kao vozač"})
		return
	}
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *MeHandler) violationList(c *gin.Context) {
	cid, ok := citizenID(c)
	if !ok {
		return
	}
	d, err := h.drivers.GetByCitizen(c.Request.Context(), cid)
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusOK, []any{}) // nije vozač → nema prekršaja
		return
	}
	if respondErr(c, err) {
		return
	}
	violations, err := h.violations.ListByDriver(c.Request.Context(), d.ID)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, violations)
}

func (h *MeHandler) fineList(c *gin.Context) {
	cid, ok := citizenID(c)
	if !ok {
		return
	}
	fines, err := h.fines.ListByCitizen(c.Request.Context(), cid)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, fines)
}
