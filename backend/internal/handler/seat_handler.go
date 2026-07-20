package handler

import (
	"errors"
	"net/http"

	"cinema-ticket-api/internal/booking"

	"github.com/gin-gonic/gin"
)

type SeatHandler struct {
	service *booking.Service
}

func NewSeatHandler(service *booking.Service) *SeatHandler {
	return &SeatHandler{
		service: service,
	}
}

type userRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

func (h *SeatHandler) ListSeats(c *gin.Context) {
	seats, err := h.service.ListSeats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not load seats",
		})
		return
	}

	c.JSON(http.StatusOK, seats)
}

func (h *SeatHandler) LockSeat(c *gin.Context) {
	var request userRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id is required",
		})
		return
	}

	bookingData, err := h.service.HoldSeat(
		c.Request.Context(),
		request.UserID,
		c.Param("seatID"),
	)
	if err != nil {
		h.writeBookingError(c, err)
		return
	}

	c.JSON(http.StatusCreated, bookingData)
}

func (h *SeatHandler) ConfirmBooking(c *gin.Context) {
	var request userRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id is required",
		})
		return
	}

	bookingData, err := h.service.ConfirmSeat(
		c.Request.Context(),
		c.Param("sessionID"),
		request.UserID,
	)
	if err != nil {
		h.writeBookingError(c, err)
		return
	}

	c.JSON(http.StatusOK, bookingData)
}

func (h *SeatHandler) ReleaseBooking(c *gin.Context) {
	var request userRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id is required",
		})
		return
	}

	err := h.service.ReleaseSeat(
		c.Request.Context(),
		c.Param("sessionID"),
		request.UserID,
	)
	if err != nil {
		h.writeBookingError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *SeatHandler) writeBookingError(c *gin.Context, err error) {
	if errors.Is(err, booking.ErrSeatUnavailable) {
		c.JSON(http.StatusConflict, gin.H{
			"error": err.Error(),
		})
		return
	}

	if errors.Is(err, booking.ErrSessionNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	if errors.Is(err, booking.ErrNotOwner) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "internal server error",
	})
}
