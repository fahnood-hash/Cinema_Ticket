package booking

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"cinema-ticket-api/internal/models"

	"cinema-ticket-api/internal/repository"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const lockDuration = 5 * time.Minute

var (
	ErrSeatUnavailable = errors.New("seat is already locked or booked")
	ErrSessionNotFound = errors.New("booking session not found or expired")
	ErrNotOwner        = errors.New("this booking belongs to another user")
)

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusConfirmed Status = "CONFIRMED"
)

type Booking struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	SeatID    string    `json:"seat_id"`
	Status    Status    `json:"status"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

type Service struct {
	redis *redis.Client
	repo  *repository.BookingRepository
}

func NewService(
	redisClient *redis.Client,
	bookingRepository *repository.BookingRepository,
) *Service {
	return &Service{
		redis: redisClient,
		repo:  bookingRepository,
	}
}
func (s *Service) HoldSeat(ctx context.Context, userID, seatID string) (*Booking, error) {
	if userID == "" {
		return nil, errors.New("user_id is required")
	}

	if seatID == "" {
		return nil, errors.New("seat_id is required")
	}

	if exists, err := s.redis.Exists(ctx, bookedKey(seatID)).Result(); err != nil {
		return nil, err
	} else if exists > 0 {
		return nil, ErrSeatUnavailable
	}

	booking := &Booking{
		ID:        primitive.NewObjectID().Hex(),
		UserID:    userID,
		SeatID:    seatID,
		Status:    StatusPending,
		ExpiresAt: time.Now().Add(lockDuration),
	}

	locked, err := s.redis.SetNX(ctx, lockKey(seatID), booking.ID, lockDuration).Result()
	if err != nil {
		return nil, err
	}

	if !locked {
		return nil, ErrSeatUnavailable
	}

	data, err := json.Marshal(booking)
	if err != nil {
		return nil, err
	}

	if err := s.redis.Set(ctx, sessionKey(booking.ID), data, lockDuration).Err(); err != nil {
		s.releaseLock(ctx, seatID, booking.ID)
		return nil, err
	}

	return booking, nil
}

func (s *Service) ConfirmSeat(ctx context.Context, sessionID, userID string) (*Booking, error) {
	booking, err := s.getSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if booking.UserID != userID {
		return nil, ErrNotOwner
	}

	lockID, err := s.redis.Get(ctx, lockKey(booking.SeatID)).Result()
	if err == redis.Nil || lockID != booking.ID {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	booked, err := s.redis.SetNX(ctx, bookedKey(booking.SeatID), booking.ID, 0).Result()
	if err != nil {
		return nil, err
	}

	if !booked {
		return nil, ErrSeatUnavailable
	}

	booking.Status = StatusConfirmed

	if err := s.repo.Create(
		ctx,
		booking.ID,
		booking.UserID,
		booking.SeatID,
	); err != nil {
		s.redis.Del(ctx, bookedKey(booking.SeatID))
		return nil, err
	}

	if err := s.redis.Del(ctx, lockKey(booking.SeatID), sessionKey(booking.ID)).Err(); err != nil {
		return nil, err
	}

	return booking, nil
}

func (s *Service) ReleaseSeat(ctx context.Context, sessionID, userID string) error {
	booking, err := s.getSession(ctx, sessionID)
	if err != nil {
		return err
	}

	if booking.UserID != userID {
		return ErrNotOwner
	}

	s.releaseLock(ctx, booking.SeatID, booking.ID)

	return s.redis.Del(ctx, sessionKey(booking.ID)).Err()
}

func (s *Service) ListSeats(ctx context.Context) ([]models.Seat, error) {
	seats := make([]models.Seat, 0, 40)

	for row := 'A'; row <= 'E'; row++ {
		for number := 1; number <= 8; number++ {
			seatID := fmt.Sprintf("%c%d", row, number)

			seat := models.Seat{
				ID:     seatID,
				Status: models.SeatAvailable,
			}

			isBooked, err := s.redis.Exists(ctx, bookedKey(seatID)).Result()
			if err != nil {
				return nil, err
			}

			if isBooked > 0 {
				seat.Status = models.SeatBooked
			} else {
				isLocked, err := s.redis.Exists(ctx, lockKey(seatID)).Result()
				if err != nil {
					return nil, err
				}

				if isLocked > 0 {
					seat.Status = models.SeatLocked
				}
			}

			seats = append(seats, seat)
		}
	}

	return seats, nil
}

func (s *Service) getSession(ctx context.Context, sessionID string) (*Booking, error) {
	data, err := s.redis.Get(ctx, sessionKey(sessionID)).Bytes()
	if err == redis.Nil {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	var booking Booking
	if err := json.Unmarshal(data, &booking); err != nil {
		return nil, err
	}

	return &booking, nil
}

func (s *Service) releaseLock(ctx context.Context, seatID, bookingID string) {
	const script = `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		end
		return 0
	`

	s.redis.Eval(ctx, script, []string{lockKey(seatID)}, bookingID)
}

func lockKey(seatID string) string {
	return "seat:lock:" + seatID
}

func bookedKey(seatID string) string {
	return "seat:booked:" + seatID
}

func sessionKey(sessionID string) string {
	return "booking:session:" + sessionID
}
