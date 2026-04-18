//go:build integration
// +build integration

package repository

import (
	"context"
	"database/sql"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var db *sql.DB

func TestMain(m *testing.M) {
	connStr := "user=test_user password=test_password dbname=test_db sslmode=disable port=5433 host=localhost"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	// wait for db
	for i := 0; i < 10; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email TEXT
	)`)
	if err != nil {
		log.Fatalf("failed to create table: %v", err)
	}

	// Clean before tests
	_, _ = db.Exec("TRUNCATE TABLE users;")

	m.Run()
}

func TestPostgresUserRepository(t *testing.T) {
	repo := NewPostgresUserRepository(db)
	ctx := context.Background()

	// 1. Create a user
	u := &User{ID: 1, Name: "Ablikon"}
	err := repo.CreateUser(ctx, u)
	require.NoError(t, err)

	// 2. Read the user
	fetched, err := repo.GetUserByID(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, u.ID, fetched.ID)
	assert.Equal(t, u.Name, fetched.Name)

	// 3. Get by Email
	fetchedByEmail, err := repo.GetByEmail(ctx, "test@test.com")
	require.NoError(t, err)
	if fetchedByEmail != nil {
		assert.Equal(t, u.ID, fetchedByEmail.ID)
		assert.Equal(t, u.Name, fetchedByEmail.Name)
	}

	// 4. Update the user
	u.Name = "Updated Ablikon"
	err = repo.UpdateUser(ctx, u)
	require.NoError(t, err)

	fetchedUpdated, err := repo.GetUserByID(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "Updated Ablikon", fetchedUpdated.Name)

	// 5. Delete the user
	err = repo.DeleteUser(ctx, 1)
	require.NoError(t, err)

	_, err = repo.GetUserByID(ctx, 1)
	require.Error(t, err)
	assert.Equal(t, "not found", err.Error())
}
