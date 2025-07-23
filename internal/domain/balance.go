package domain

import "time"

// Balance represents a user's account balance.
type Balance struct {
	UserID        int
	Amount        float64
	LastUpdatedAt time.Time
}
