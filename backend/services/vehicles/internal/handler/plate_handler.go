package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"euprava/vehicles/internal/service"
)

// PlateHandler izlaže /plate-reservations rute — zahtevi za personalizovane
// registarske tablice.
type PlateHandler struct {
	plates *service.PlateService
}

func NewPlateHandler(plates *service.PlateService) *PlateHandler {
	return &PlateHandler{plates: plates}
}

func (h *PlateHandler) RegisterRoutes(r gin.IRouter) {
	g := r.Group("/plate-reservations")
	g.GET("", h.list)
	g.POST("", h.request)
	g.PUT("/:id/decision", h.decide)
}

// request podnosi građanin — traži personalizovanu tablicu za sebe.
func (h *PlateHandler) request(c *gin.Context) {
	cid, _, ok := requester(c)
	if !ok {
		return
	}
	var in service.RequestPlateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "neispravni podaci: " + err.Error()})
		return
	}
	r, err := h.plates.Request(c.Request.Context(), cid, in)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusCreated, r)
}

// decide odobrava/odbija zahtev — poziva ga službenik.
func (h *PlateHandler) decide(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	var in struct {
		Approve bool `json:"approve"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "neispravni podaci: " + err.Error()})
		return
	}
	r, err := h.plates.Decide(c.Request.Context(), id, in.Approve)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, r)
}

// list vraća sve zahteve — pregleda ih službenik pri obradi reda čekanja.
func (h *PlateHandler) list(c *gin.Context) {
	reservations, err := h.plates.List(c.Request.Context())
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, reservations)
}
