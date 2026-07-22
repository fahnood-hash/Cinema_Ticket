package repository

import (
	"context"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BookingRepository struct {
	collection *mongo.Collection
}
type BookingRecord struct {
	ID        string    `bson:"_id" json:"id"`
	UserID    string    `bson:"user_id" json:"user_id"`
	SeatID    string    `bson:"seat_id" json:"seat_id"`
	Status    string    `bson:"status" json:"status"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
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

func (r *BookingRepository) ListAll(
	ctx context.Context,
	userID string,
	seatID string,
) ([]BookingRecord, error) {
	filter := bson.M{}

	if userID != "" {
		filter["user_id"] = bson.M{
			"$regex":   regexp.QuoteMeta(userID),
			"$options": "i",
		}
	}

	if seatID != "" {
		filter["seat_id"] = bson.M{
			"$regex":   regexp.QuoteMeta(seatID),
			"$options": "i",
		}
	}

	cursor, err := r.collection.Find(
		ctx,
		filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	bookings := []BookingRecord{}

	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, err
	}

	return bookings, nil
}
