package service

import (
	"context"
	"errors"
	"fmt"
	"practice-8/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUserByID(ctx context.Context, id int) (*repository.User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *UserService) CreateUser(ctx context.Context, user *repository.User) error {
	return s.repo.CreateUser(ctx, user)
}

func (s *UserService) RegisterUser(ctx context.Context, user *repository.User, email string) error {
	existing, err := s.repo.GetByEmail(ctx, email)
	if existing != nil {
		return fmt.Errorf("user with this email already exists")
	}
	if err != nil {
		return fmt.Errorf("error getting user with this email")
	}
	return s.repo.CreateUser(ctx, user)
}

func (s *UserService) UpdateUserName(ctx context.Context, id int, newName string) error {
	if newName == "" {
		return fmt.Errorf("name cannot be empty")
	}
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return err
	}
	user.Name = newName
	return s.repo.UpdateUser(ctx, user)
}

func (s *UserService) DeleteUser(ctx context.Context, id int) error {
	if id == 1 {
		return errors.New("it is not allowed to delete admin user")
	}
	return s.repo.DeleteUser(ctx, id)
}
