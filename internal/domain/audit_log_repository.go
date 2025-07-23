package domain

// AuditLogRepository defines methods for audit log data access.
type AuditLogRepository interface {
	Create(log *AuditLog) error
	ListByEntity(entityType string, entityID int) ([]*AuditLog, error)
}
