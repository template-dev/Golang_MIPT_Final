package pg

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
}

func CreateUser(ctx context.Context, db *sql.DB, id uuid.UUID, email string, passwordHash string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return errors.New("email is empty")
	}
	if passwordHash == "" {
		return errors.New("password hash is empty")
	}

	_, err := db.ExecContext(ctx,
		`INSERT INTO users(id, email, password_hash) VALUES($1,$2,$3)`,
		id, email, passwordHash,
	)
	return err
}

func GetUserByEmail(ctx context.Context, db *sql.DB, email string) (User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	var u User
	err := db.QueryRowContext(ctx,
		`SELECT id, email, password_hash FROM users WHERE email=$1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash)
	return u, err
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return true
		}
	}
	return false
}
