package models

type SeatStatus string

const (
	SeatAvailable SeatStatus = "AVAILABLE"
	SeatLocked    SeatStatus = "LOCKED"
	SeatBooked    SeatStatus = "BOOKED"
)

type Seat struct {
	ID       string     `json:"id" bson:"id"`
	Status   SeatStatus `json:"status" bson:"status"`
	LockedBy string     `json:"locked_by,omitempty" bson:"locked_by,omitempty"`
}
