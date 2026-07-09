package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"euprava/citizen/internal/service"
)

// UserHandler izlaže /users rute za administratora
// (gateway propušta samo ulogu admin).
type UserHandler struct {
	auth     *service.AuthService
	citizens *service.CitizenService
}

func NewUserHandler(auth *service.AuthService, citizens *service.CitizenService) *UserHandler {
	return &UserHandler{auth: auth, citizens: citizens}
}

func (h *UserHandler) RegisterRoutes(r gin.IRouter) {
	g := r.Group("/users")
	g.GET("", h.list)
	g.POST("/officers", h.createOfficer)
}

// list vraća sve korisnike sistema.
func (h *UserHandler) list(c *gin.Context) {
	users, err := h.citizens.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// createOfficer kreira nalog policajca — jedini način da nastane novi
// policajac kroz sistem (javna registracija daje samo ulogu citizen).
func (h *UserHandler) createOfficer(c *gin.Context) {
	var in service.RegisterInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "neispravni podaci: " + err.Error()})
		return
	}
	u, err := h.auth.CreateOfficer(c.Request.Context(), in)
	if errors.Is(err, service.ErrEmailTaken) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, u)
}
