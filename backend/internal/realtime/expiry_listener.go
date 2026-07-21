package realtime

import (
	"context"
	"log"
	"strings"
	"time"

	"cinema-ticket-api/internal/queue"

	"github.com/redis/go-redis/v9"
)

func StartExpiredLockListener(
	ctx context.Context,
	redisClient *redis.Client,
	hub *Hub,
	publisher *queue.Publisher,
) {
	pubsub := redisClient.PSubscribe(ctx, "__keyevent@0__:expired")
	defer pubsub.Close()

	if _, err := pubsub.Receive(ctx); err != nil {
		log.Printf("Redis expiry-listener subscription failed: %v", err)
		return
	}

	log.Println("Listening for expired Redis seat locks")

	messages := pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return

		case message, ok := <-messages:
			if !ok {
				return
			}

			if !strings.HasPrefix(message.Payload, "seat:lock:") {
				continue
			}

			seatID := strings.TrimPrefix(message.Payload, "seat:lock:")

			hub.Broadcast(SeatEvent{
				Type:   "seat.updated",
				SeatID: seatID,
				Status: "AVAILABLE",
			})

			err := publisher.PublishBookingConfirmed(
				ctx,
				queue.BookingConfirmedEvent{
					EventType:  "BOOKING_TIMEOUT",
					SeatID:     seatID,
					OccurredAt: time.Now(),
				},
			)
			if err != nil {
				log.Printf("Timeout audit event failed: %v", err)
			}
		}
	}
}
