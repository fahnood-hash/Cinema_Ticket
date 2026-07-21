package queue

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const BookingEventsQueue = "booking.events"

type BookingConfirmedEvent struct {
	EventType  string    `json:"event_type"`
	BookingID  string    `json:"booking_id"`
	UserID     string    `json:"user_id"`
	SeatID     string    `json:"seat_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

type Publisher struct {
	connection *amqp.Connection
}

func NewPublisher(url string) (*Publisher, error) {
	connection, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		connection: connection,
	}, nil
}

func (p *Publisher) Close() error {
	return p.connection.Close()
}

func (p *Publisher) PublishBookingConfirmed(
	ctx context.Context,
	event BookingConfirmedEvent,
) error {
	channel, err := p.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	_, err = channel.QueueDeclare(
		BookingEventsQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return channel.PublishWithContext(
		ctx,
		"",
		BookingEventsQueue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         body,
		},
	)
}
