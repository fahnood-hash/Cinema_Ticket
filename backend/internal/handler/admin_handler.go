package handler

import (
	"net/http"

	"cinema-ticket-api/internal/repository"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	repo *repository.BookingRepository
}

func NewAdminHandler(repo *repository.BookingRepository) *AdminHandler {
	return &AdminHandler{
		repo: repo,
	}
}

func (h *AdminHandler) ListBookings(c *gin.Context) {
	bookings, err := h.repo.ListAll(
		c.Request.Context(),
		c.Query("user_id"),
		c.Query("seat_id"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not load bookings",
		})
		return
	}

	c.JSON(http.StatusOK, bookings)
}
