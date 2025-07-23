package domain

import "time"

// AuditLog represents an audit log entry for tracking changes.
type AuditLog struct {
	ID         int
	EntityType string
	EntityID   int
	Action     string
	Details    string
	CreatedAt  time.Time
}
