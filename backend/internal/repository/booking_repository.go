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

func (r *BookingRepository) ListConfirmedSeatIDs(
	ctx context.Context,
) (map[string]bool, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"status": "CONFIRMED",
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	bookedSeats := make(map[string]bool)

	for cursor.Next(ctx) {
		var booking struct {
			SeatID string `bson:"seat_id"`
		}

		if err := cursor.Decode(&booking); err != nil {
			return nil, err
		}

		bookedSeats[booking.SeatID] = true
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return bookedSeats, nil
}
