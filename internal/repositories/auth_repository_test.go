package repositories

import (
	"testing"

	"gocart/internal/models"

	"gorm.io/gorm"
)

func TestAuthRepository_UpdatePassword(t *testing.T) {
	db := openTestDB(t)
	users := seedUsers(t, db, 1)
	repo := NewAuthRepository(db)

	t.Run("updates the stored hash", func(t *testing.T) {
		u := &models.User{}
		if err := u.HashPassword("secret123"); err != nil {
			t.Fatalf("hash: %v", err)
		}

		if err := repo.UpdatePassword(users[0].ID, u.Password); err != nil {
			t.Fatalf("UpdatePassword: %v", err)
		}

		var reloaded models.User
		if err := db.First(&reloaded, users[0].ID).Error; err != nil {
			t.Fatalf("reload user: %v", err)
		}
		if !reloaded.VerifyPassword("secret123") {
			t.Fatal("want stored hash to verify against the new password")
		}
	})

	t.Run("missing user", func(t *testing.T) {
		u := &models.User{}
		if err := u.HashPassword("secret123"); err != nil {
			t.Fatalf("hash: %v", err)
		}

		if err := repo.UpdatePassword(99999, u.Password); err != gorm.ErrRecordNotFound {
			t.Fatalf("want ErrRecordNotFound, got %v", err)
		}
	})
}
