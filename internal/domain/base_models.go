package domain

import "time"

// EntityWithID base interface for all domain entities
type EntityWithID interface {
	GetID() uint
	SetID(uint)
}

// BaseEntity provides common fields for most domain entities
type BaseEntity struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// GetID returns the entity ID
func (b *BaseEntity) GetID() uint {
	return b.ID
}

// SetID sets the entity ID
func (b *BaseEntity) SetID(id uint) {
	b.ID = id
}

// BaseEntityWithSoftDelete adds soft delete capability
type BaseEntityWithSoftDelete struct {
	BaseEntity
	DeletedAt *time.Time `gorm:"index" json:"-"`
}

// IsDeleted checks if entity is soft deleted
func (b *BaseEntityWithSoftDelete) IsDeleted() bool {
	return b.DeletedAt != nil
}

// AuditableEntity adds audit tracking
type AuditableEntity struct {
	BaseEntity
	CreatedBy uint `json:"created_by,omitempty"`
	UpdatedBy uint `json:"updated_by,omitempty"`
}

// TimestampOnlyEntity for entities that only need timestamps (no ID management)
type TimestampOnlyEntity struct {
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
