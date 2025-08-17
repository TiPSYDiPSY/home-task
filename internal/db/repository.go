package db

import (
	"context"
)

type UserRepository interface {
	GetUserData(ctx context.Context, userID uint64) (User, error)
}

func (r *PostgresDBDataStore) GetUserData(ctx context.Context, userID uint64) (user User, err error) {
	return user, r.db.WithContext(ctx).First(&user, userID).Error
}
