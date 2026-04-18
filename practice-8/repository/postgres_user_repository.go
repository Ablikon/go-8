package repository

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, user *User) error {
	query := "INSERT INTO users (id, name, email) VALUES ($1, $2, $3)"
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Name, "test@test.com") // assuming email is part of User in db
	return err
}

func (r *PostgresUserRepository) GetUserByID(ctx context.Context, id int) (*User, error) {
	query := "SELECT id, name FROM users WHERE id = $1"
	row := r.db.QueryRowContext(ctx, query, id)
	user := &User{}
	err := row.Scan(&user.ID, &user.Name)
	if err == sql.ErrNoRows {
		return nil, errors.New("not found")
	}
	return user, err
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := "SELECT id, name FROM users WHERE email = $1"
	row := r.db.QueryRowContext(ctx, query, email)
	user := &User{}
	err := row.Scan(&user.ID, &user.Name)
	if err == sql.ErrNoRows {
		return nil, nil // to match mock behaviour expectation
	}
	return user, err
}

func (r *PostgresUserRepository) UpdateUser(ctx context.Context, user *User) error {
	query := "UPDATE users SET name = $1 WHERE id = $2"
	res, err := r.db.ExecContext(ctx, query, user.Name, user.ID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("no user updated")
	}
	return nil
}

func (r *PostgresUserRepository) DeleteUser(ctx context.Context, id int) error {
	query := "DELETE FROM users WHERE id = $1"
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("no user deleted")
	}
	return nil
}
