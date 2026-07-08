package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"euprava/traffic-police/internal/service"
)

// FineHandler izlaže /fines rute.
type FineHandler struct {
	fines *service.FineService
}

func NewFineHandler(fines *service.FineService) *FineHandler {
	return &FineHandler{fines: fines}
}

func (h *FineHandler) RegisterRoutes(r gin.IRouter) {
	g := r.Group("/fines")
	g.GET("", h.list)
	g.GET("/:id", h.get)
	g.PUT("/:id/pay", h.pay)
}

func (h *FineHandler) list(c *gin.Context) {
	fines, err := h.fines.List(c.Request.Context())
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, fines)
}

func (h *FineHandler) get(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	fd, err := h.fines.Get(c.Request.Context(), id)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, fd)
}

// pay je interni endpoint — poziva ga Payment servis po uspešnom plaćanju
// (gateway spolja ne propušta PUT na /api/fines).
func (h *FineHandler) pay(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	fd, err := h.fines.Pay(c.Request.Context(), id)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, fd)
}
