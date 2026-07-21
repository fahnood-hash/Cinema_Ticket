package main

import (
	"cinema-ticket-api/internal/booking"
	"cinema-ticket-api/internal/database"
	"cinema-ticket-api/internal/handler"
	"cinema-ticket-api/internal/queue"
	"cinema-ticket-api/internal/realtime"
	"cinema-ticket-api/internal/repository"
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
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

	redisClient, err := database.ConnectRedis(
		getEnv("REDIS_ADDR", "redis:6379"),
	)
	if err != nil {
		log.Fatal("Redis connection failed: ", err)
	}
	defer redisClient.Close()

	publisher, err := queue.NewPublisher(
		getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
	)
	if err != nil {
		log.Fatal("RabbitMQ connection failed: ", err)
	}
	defer publisher.Close()

	r := gin.Default()
	r.SetTrustedProxies(nil)
	seatHub := realtime.NewHub()
	r.GET("/ws", seatHub.HandleConnection)

	bookingRepo := repository.NewBookingRepository(
		mongoClient.Database("cinema_booking"),
	)

	bookingService := booking.NewService(
		redisClient,
		bookingRepo,
		publisher,
		seatHub,
	)

	seatHandler := handler.NewSeatHandler(bookingService)

	r.GET("/seats", seatHandler.ListSeats)
	r.POST("/seats/:seatID/lock", seatHandler.LockSeat)
	r.POST("/bookings/:sessionID/confirm", seatHandler.ConfirmBooking)
	r.DELETE("/bookings/:sessionID", seatHandler.ReleaseBooking)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "OK",
			"mongo":  "connected",
			"redis":  "connected",
		})
	})

	r.Run(":8080")
}
