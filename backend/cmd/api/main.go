package main

import (
	"cinema-ticket-api/internal/authentication"
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

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
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

	firebaseAuth, err := authentication.NewFirebaseAuthenticator(
		getEnv(
			"FIREBASE_CREDENTIALS_PATH",
			"secrets/firebase-service-account.json",
		),
	)
	if err != nil {
		log.Fatal("Firebase initialization failed: ", err)
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Use(corsMiddleware())

	seatHub := realtime.NewHub()
	r.GET("/ws", seatHub.HandleConnection)
	r.GET("/me", firebaseAuth.RequireAuth(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"user_id": authentication.UserID(c),
			"email":   c.GetString("email"),
		})
	})

	go realtime.StartExpiredLockListener(
		context.Background(),
		redisClient,
		seatHub,
		publisher,
	)

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
