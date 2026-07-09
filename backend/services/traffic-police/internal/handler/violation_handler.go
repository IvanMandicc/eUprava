package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"euprava/traffic-police/internal/model"
	"euprava/traffic-police/internal/service"
)

// ViolationHandler izlaže /violations rute i šifarnik tipova prekršaja.
type ViolationHandler struct {
	violations *service.ViolationService
}

func NewViolationHandler(violations *service.ViolationService) *ViolationHandler {
	return &ViolationHandler{violations: violations}
}

func (h *ViolationHandler) RegisterRoutes(r gin.IRouter) {
	g := r.Group("/violations")
	g.GET("", h.list)
	g.POST("", h.create)
	g.PUT("/:id", h.update)
	g.DELETE("/:id", h.delete)
	r.GET("/violation-types", h.types)
	// Open data: javna, anonimna statistika — gateway je pušta bez tokena.
	r.GET("/open-data/violation-stats", h.stats)
}

func (h *ViolationHandler) stats(c *gin.Context) {
	stats, err := h.violations.Stats(c.Request.Context())
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *ViolationHandler) list(c *gin.Context) {
	violations, err := h.violations.List(c.Request.Context())
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, violations)
}

func (h *ViolationHandler) create(c *gin.Context) {
	var in service.CreateViolationInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "neispravni podaci: " + err.Error()})
		return
	}
	v, err := h.violations.Create(c.Request.Context(), in)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusCreated, v)
}

func (h *ViolationHandler) update(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	var in service.UpdateViolationInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "neispravni podaci: " + err.Error()})
		return
	}
	v, err := h.violations.Update(c.Request.Context(), id, in)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *ViolationHandler) delete(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	if respondErr(c, h.violations.Delete(c.Request.Context(), id)) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "prekršaj je obrisan"})
}

func (h *ViolationHandler) types(c *gin.Context) {
	c.JSON(http.StatusOK, model.ViolationCatalog)
}
