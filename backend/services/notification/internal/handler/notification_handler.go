package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"euprava/notification/internal/repository"
	"euprava/notification/internal/service"
)

// NotificationHandler izlaže /notifications rute.
type NotificationHandler struct {
	notifications *service.NotificationService
}

func NewNotificationHandler(notifications *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifications: notifications}
}

func (h *NotificationHandler) RegisterRoutes(r gin.IRouter) {
	g := r.Group("/notifications")
	g.POST("", h.create)
	g.GET("", h.list)
	g.PUT("/:id/read", h.markRead)
}

// create pozivaju drugi servisi (interno); gateway blokira POST spolja.
func (h *NotificationHandler) create(c *gin.Context) {
	var in service.CreateNotificationInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "neispravni podaci: " + err.Error()})
		return
	}
	n, err := h.notifications.Create(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, n)
}

// list vraća obaveštenja ulogovanog građanina (X-User-Id postavlja gateway).
func (h *NotificationHandler) list(c *gin.Context) {
	cid, ok := citizenID(c)
	if !ok {
		return
	}
	notifications, err := h.notifications.ListByCitizen(c.Request.Context(), cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, notifications)
}

func (h *NotificationHandler) markRead(c *gin.Context) {
	cid, ok := citizenID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "neispravan id"})
		return
	}
	err = h.notifications.MarkRead(c.Request.Context(), id, cid)
	if errors.Is(err, repository.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "obaveštenje nije pronađeno"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "obaveštenje je označeno kao pročitano"})
}

func citizenID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.GetHeader("X-User-Id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "nedostaje X-User-Id header"})
		return 0, false
	}
	return id, true
}
