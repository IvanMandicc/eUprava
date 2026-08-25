package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"euprava/vehicles/internal/service"
)

// MeHandler izlaže /me rute — građanin pregleda sopstvena vozila i zahteve
// za personalizovane tablice. Identitet stiže kroz X-User-Id header koji
// postavlja API Gateway.
type MeHandler struct {
	vehicles *service.VehicleService
	plates   *service.PlateService
}

func NewMeHandler(vehicles *service.VehicleService, plates *service.PlateService) *MeHandler {
	return &MeHandler{vehicles: vehicles, plates: plates}
}

func (h *MeHandler) RegisterRoutes(r gin.IRouter) {
	g := r.Group("/me")
	g.GET("/vehicles", h.myVehicles)
	g.GET("/plate-reservations", h.myReservations)
}

func (h *MeHandler) myVehicles(c *gin.Context) {
	cid, _, ok := requester(c)
	if !ok {
		return
	}
	vehicles, err := h.vehicles.ListByOwner(c.Request.Context(), cid)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, vehicles)
}

func (h *MeHandler) myReservations(c *gin.Context) {
	cid, _, ok := requester(c)
	if !ok {
		return
	}
	reservations, err := h.plates.ListByRequester(c.Request.Context(), cid)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, reservations)
}
