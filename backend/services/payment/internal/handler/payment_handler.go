package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"euprava/payment/internal/repository"
	"euprava/payment/internal/service"
)

// PaymentHandler izlaže /payments rute.
type PaymentHandler struct {
	payments *service.PaymentService
}

func NewPaymentHandler(payments *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{payments: payments}
}

func (h *PaymentHandler) RegisterRoutes(r gin.IRouter) {
	g := r.Group("/payments")
	g.POST("", h.create)
	g.POST("/:id/confirm", h.confirm)
	g.GET("", h.list)
}

type createInput struct {
	FineID int64 `json:"fineId" binding:"required"`
}

func (h *PaymentHandler) create(c *gin.Context) {
	cid, ok := citizenID(c)
	if !ok {
		return
	}
	var in createInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "neispravni podaci: " + err.Error()})
		return
	}
	p, err := h.payments.Create(c.Request.Context(), cid, in.FineID)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *PaymentHandler) confirm(c *gin.Context) {
	cid, ok := citizenID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "neispravan id"})
		return
	}
	p, err := h.payments.Confirm(c.Request.Context(), id, cid)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *PaymentHandler) list(c *gin.Context) {
	cid, ok := citizenID(c)
	if !ok {
		return
	}
	payments, err := h.payments.ListByCitizen(c.Request.Context(), cid)
	if respondErr(c, err) {
		return
	}
	c.JSON(http.StatusOK, payments)
}

func citizenID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.GetHeader("X-User-Id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "nedostaje X-User-Id header"})
		return 0, false
	}
	return id, true
}

func respondErr(c *gin.Context, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, repository.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "zapis nije pronađen"})
	case errors.Is(err, service.ErrNotYourFine), errors.Is(err, service.ErrNotYourPayment):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrFineAlreadyPaid), errors.Is(err, service.ErrPaymentExists):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	return true
}
