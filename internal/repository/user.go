package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/firmanains/cinema-ticketing/internal/domain"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, "SELECT * FROM users WHERE email = $1", email)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
		VALUES (:id, :name, :email, :password_hash, :created_at, :updated_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return errors.New("email already registered")
		}
		return err
	}
	return nil
}
