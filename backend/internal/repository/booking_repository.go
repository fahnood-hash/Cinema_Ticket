package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type BookingRepository struct {
	collection *mongo.Collection
}

func NewBookingRepository(database *mongo.Database) *BookingRepository {
	return &BookingRepository{
		collection: database.Collection("bookings"),
	}
}

func (r *BookingRepository) Create(
	ctx context.Context,
	bookingID string,
	userID string,
	seatID string,
) error {
	_, err := r.collection.InsertOne(ctx, bson.M{
		"_id":        bookingID,
		"user_id":    userID,
		"seat_id":    seatID,
		"status":     "CONFIRMED",
		"created_at": time.Now(),
	})

	return err
}
