package repository

import "context"

type User struct {
	ID   int
	Name string
}

type UserRepository interface {
	GetUserByID(ctx context.Context, id int) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	CreateUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id int) error
}
