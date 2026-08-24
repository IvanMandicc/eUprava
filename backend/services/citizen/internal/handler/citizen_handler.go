package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"euprava/citizen/internal/repository"
	"euprava/citizen/internal/service"
)

// CitizenHandler izlaže /citizens rute.
type CitizenHandler struct {
	citizens *service.CitizenService
}

func NewCitizenHandler(citizens *service.CitizenService) *CitizenHandler {
	return &CitizenHandler{citizens: citizens}
}

func (h *CitizenHandler) RegisterRoutes(r gin.IRouter) {
	g := r.Group("/citizens")
	g.GET("/me", h.me)
	g.GET("/search", h.search)
	g.GET("/:id", h.getByID)
}

// search pronalazi građanina po JMBG-u (?jmbg=...) — koristi ga policajac
// pri evidentiranju vozača, umesto unosa internog ID-ja iz baze.
func (h *CitizenHandler) search(c *gin.Context) {
	jmbg := c.Query("jmbg")
	if len(jmbg) != 13 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JMBG mora imati tačno 13 cifara"})
		return
	}
	u, err := h.citizens.SearchByJMBG(c.Request.Context(), jmbg)
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "građanin sa unetim JMBG-om nije pronađen"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, u)
}

// me vraća profil ulogovanog korisnika; gateway prosleđuje X-User-Id header.
func (h *CitizenHandler) me(c *gin.Context) {
	id, err := strconv.ParseInt(c.GetHeader("X-User-Id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "nedostaje X-User-Id header"})
		return
	}
	h.respondUser(c, id)
}

func (h *CitizenHandler) getByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "neispravan id"})
		return
	}
	h.respondUser(c, id)
}

func (h *CitizenHandler) respondUser(c *gin.Context, id int64) {
	u, err := h.citizens.GetByID(c.Request.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "građanin nije pronađen"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, u)
}
