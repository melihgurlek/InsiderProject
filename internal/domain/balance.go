package domain

import (
	"sync"
	"time"
)

// Balance represents a user's account balance with thread-safe operations.
type Balance struct {
	UserID        int
	Amount        float64
	LastUpdatedAt time.Time
	mu            sync.RWMutex // protects Amount and LastUpdatedAt
}

// NewBalance creates a new Balance instance
func NewBalance(userID int, amount float64) *Balance {
	return &Balance{
		UserID:        userID,
		Amount:        amount,
		LastUpdatedAt: time.Now(),
	}
}

// GetAmount returns the current balance amount in a thread-safe manner
func (b *Balance) GetAmount() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Amount
}

// SetAmount sets the balance amount in a thread-safe manner
func (b *Balance) SetAmount(amount float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Amount = amount
	b.LastUpdatedAt = time.Now()
}

// AddAmount adds to the balance in a thread-safe manner
func (b *Balance) AddAmount(amount float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Amount += amount
	b.LastUpdatedAt = time.Now()
}

// SubtractAmount subtracts from the balance in a thread-safe manner
// Returns false if insufficient funds
func (b *Balance) SubtractAmount(amount float64) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.Amount < amount {
		return false
	}
	b.Amount -= amount
	b.LastUpdatedAt = time.Now()
	return true
}

// GetLastUpdatedAt returns the last update time in a thread-safe manner
func (b *Balance) GetLastUpdatedAt() time.Time {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.LastUpdatedAt
}
