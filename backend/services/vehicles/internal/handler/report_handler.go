package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"euprava/vehicles/internal/repository"
	"euprava/vehicles/internal/service"
)

// ReportHandler izlaže generisanje izveštaja o vozilu i javnu, neprijavljenu
// proveru njegove autentičnosti po verifikacionom kodu.
type ReportHandler struct {
	reports *service.ReportService
}

func NewReportHandler(reports *service.ReportService) *ReportHandler {
	return &ReportHandler{reports: reports}
}

func (h *ReportHandler) RegisterRoutes(r gin.IRouter) {
	r.POST("/vehicles/:id/reports", h.generate)
	r.GET("/reports/verify/:code", h.verify)
}

// generate kreira izveštaj — dostupno vlasniku vozila ili službeniku.
func (h *ReportHandler) generate(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	cid, role, ok := requester(c)
	if !ok {
		return
	}
	details, err := h.reports.Generate(c.Request.Context(), id, cid, role)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusCreated, details)
}

// verify je javna ruta (gateway je pušta bez tokena) — bilo ko sa kodom
// izveštaja može proveriti njegovu autentičnost.
func (h *ReportHandler) verify(c *gin.Context) {
	code := c.Param("code")
	details, err := h.reports.Verify(c.Request.Context(), code)
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"valid": false, "error": "izveštaj sa ovim kodom ne postoji"})
		return
	}
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, gin.H{"valid": true, "report": details})
}
