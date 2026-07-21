package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuditLogRepository struct {
	collection *mongo.Collection
}

func NewAuditLogRepository(database *mongo.Database) *AuditLogRepository {
	return &AuditLogRepository{
		collection: database.Collection("audit_logs"),
	}
}

func (r *AuditLogRepository) Create(
	ctx context.Context,
	eventType string,
	bookingID string,
	userID string,
	seatID string,
) error {
	_, err := r.collection.InsertOne(ctx, bson.M{
		"event_type": eventType,
		"booking_id": bookingID,
		"user_id":    userID,
		"seat_id":    seatID,
		"created_at": time.Now(),
	})

	return err
}
