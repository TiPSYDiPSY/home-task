package db

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uint64 `gorm:"primaryKey"`
	Balance   int64  `gorm:"not null;default:0;check:balance >= 0"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Transactions []Transaction `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

type Transaction struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID        uint64    `gorm:"not null"`
	Amount        int64     `gorm:"not null"`
	State         string    `gorm:"type:varchar(4);not null"`
	SourceType    string    `gorm:"type:varchar(10);not null"`
	TransactionID string    `gorm:"uniqueIndex;not null"`
	ProcessedAt   time.Time `gorm:"not null;default:now()"`
}
