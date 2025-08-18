package db

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
)

func (r *PostgresDBDataStore) RunAutoMigrate(ctx context.Context) error {
	log.WithContext(ctx).Info("auto-migration started")

	if err := r.db.WithContext(ctx).AutoMigrate(
		&User{},
		&Transaction{},
	); err != nil {
		return fmt.Errorf("auto-migration failed: %w", err)
	}

	log.WithContext(ctx).Info("auto-migration of tables finished")

	log.WithContext(ctx).Info("Setting up predefined users...")

	//nolint: revive,mnd // This is stub data
	predefinedUsers := []*User{
		{ID: 1, Balance: 0},
		{ID: 2, Balance: 0},
		{ID: 3, Balance: 0},
	}

	for _, user := range predefinedUsers {
		if err := r.db.WithContext(ctx).FirstOrCreate(user, User{ID: user.ID}).Error; err != nil {
			return fmt.Errorf("failed to create predefined user with ID %d: %w", user.ID, err)
		}
	}

	return nil
}
