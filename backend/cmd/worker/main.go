package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"cinema-ticket-api/internal/database"
	"cinema-ticket-api/internal/queue"
	"cinema-ticket-api/internal/repository"

	amqp "github.com/rabbitmq/amqp091-go"
)

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func main() {
	mongoClient, err := database.ConnectMongo(
		getEnv("MONGO_URI", "mongodb://mongo:27017"),
	)
	if err != nil {
		log.Fatal("MongoDB connection failed: ", err)
	}
	defer mongoClient.Disconnect(context.Background())

	auditLogRepo := repository.NewAuditLogRepository(
		mongoClient.Database("cinema_booking"),
	)

	rabbitURL := getEnv(
		"RABBITMQ_URL",
		"amqp://guest:guest@rabbitmq:5672/",
	)

	for {
		runWorker(rabbitURL, auditLogRepo)
		log.Println("RabbitMQ connection closed; retrying in 3 seconds")
		time.Sleep(3 * time.Second)
	}
}

func runWorker(rabbitURL string, auditLogRepo *repository.AuditLogRepository) {
	connection, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Println("RabbitMQ connection failed: ", err)
		return
	}
	defer connection.Close()

	channel, err := connection.Channel()
	if err != nil {
		log.Println("RabbitMQ channel failed: ", err)
		return
	}
	defer channel.Close()

	_, err = channel.QueueDeclare(
		queue.BookingEventsQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Println("RabbitMQ queue declaration failed: ", err)
		return
	}

	messages, err := channel.Consume(
		queue.BookingEventsQueue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Println("RabbitMQ consumer failed: ", err)
		return
	}

	log.Println("Audit-log worker is waiting for booking events")

	for message := range messages {
		var event queue.BookingConfirmedEvent

		if err := json.Unmarshal(message.Body, &event); err != nil {
			log.Println("Invalid RabbitMQ message: ", err)
			message.Nack(false, false)
			continue
		}

		err := auditLogRepo.Create(
			context.Background(),
			event.EventType,
			event.BookingID,
			event.UserID,
			event.SeatID,
		)
		if err != nil {
			log.Println("Audit-log save failed: ", err)
			message.Nack(false, true)
			continue
		}

		message.Ack(false)
		log.Println("Audit log saved for booking: ", event.BookingID)
	}
}
