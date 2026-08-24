package handler

import (
	"net/http"
	"time"

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

// stats vraća statistiku, opciono filtriranu upit parametrima
// ?from=GGGG-MM-DD&to=GGGG-MM-DD (oba opciona, granice datuma prekršaja).
func (h *ViolationHandler) stats(c *gin.Context) {
	from, err := parseDateParam(c.Query("from"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "neispravan format datuma 'from' (očekivano GGGG-MM-DD)"})
		return
	}
	to, err := parseDateParam(c.Query("to"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "neispravan format datuma 'to' (očekivano GGGG-MM-DD)"})
		return
	}
	if to != nil {
		end := to.Add(24 * time.Hour) // uključi ceo dan "to"
		to = &end
	}

	stats, err := h.violations.Stats(c.Request.Context(), from, to)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, stats)
}

func parseDateParam(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, err
	}
	return &t, nil
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
