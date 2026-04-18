package service

import (
	"context"
	"errors"
	"practice-8/repository"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetUserByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		userService := NewUserService(mockRepo)

		user := &repository.User{ID: 1, Name: "Bakytzhan Agai"}
		ctx := context.Background()

		mockRepo.EXPECT().GetUserByID(ctx, 1).Return(user, nil)

		result, err := userService.GetUserByID(ctx, 1)

		assert.NoError(t, err)
		assert.Equal(t, user, result)
	})

	t.Run("timeout context", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		userService := NewUserService(mockRepo)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		time.Sleep(2 * time.Millisecond) // Ensure timeout

		mockRepo.EXPECT().GetUserByID(ctx, 1).Return(nil, context.DeadlineExceeded)

		result, err := userService.GetUserByID(ctx, 1)

		assert.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)
		assert.Nil(t, result)
	})

	t.Run("cancelled context", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := repository.NewMockUserRepository(ctrl)
		userService := NewUserService(mockRepo)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		mockRepo.EXPECT().GetUserByID(ctx, 1).Return(nil, context.Canceled)

		result, err := userService.GetUserByID(ctx, 1)

		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
		assert.Nil(t, result)
	})
}

func TestRegisterUser(t *testing.T) {
	tests := []struct {
		name        string
		user        *repository.User
		email       string
		mockSetup   func(mockRepo *repository.MockUserRepository, ctx context.Context, user *repository.User, email string)
		expectError bool
		errorMsg    string
	}{
		{
			name:  "User already exists",
			user:  &repository.User{ID: 2, Name: "New User"},
			email: "test@test.com",
			mockSetup: func(mockRepo *repository.MockUserRepository, ctx context.Context, user *repository.User, email string) {
				mockRepo.EXPECT().GetByEmail(ctx, email).Return(&repository.User{ID: 1}, nil)
			},
			expectError: true,
			errorMsg:    "user with this email already exists",
		},
		{
			name:  "New User -> Success",
			user:  &repository.User{ID: 2, Name: "New User"},
			email: "test@test.com",
			mockSetup: func(mockRepo *repository.MockUserRepository, ctx context.Context, user *repository.User, email string) {
				// GetByEmail needs to return an error, but actually in code it checks: if existing != nil ... if err != nil. Wait, the code says: return fmt.Errorf("error getting user with this email").
				// This means GetByEmail must return nil, nil to proceed to Creation!
				mockRepo.EXPECT().GetByEmail(ctx, email).Return(nil, nil)
				mockRepo.EXPECT().CreateUser(ctx, user).Return(nil)
			},
			expectError: false,
		},
		{
			name:  "Repository error on CreateUser",
			user:  &repository.User{ID: 2, Name: "New User"},
			email: "test@test.com",
			mockSetup: func(mockRepo *repository.MockUserRepository, ctx context.Context, user *repository.User, email string) {
				mockRepo.EXPECT().GetByEmail(ctx, email).Return(nil, nil)
				mockRepo.EXPECT().CreateUser(ctx, user).Return(errors.New("db error"))
			},
			expectError: true,
			errorMsg:    "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRepo := repository.NewMockUserRepository(ctrl)
			userService := NewUserService(mockRepo)
			ctx := context.Background()

			tt.mockSetup(mockRepo, ctx, tt.user, tt.email)
			err := userService.RegisterUser(ctx, tt.user, tt.email)

			if tt.expectError {
				require.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUpdateUserName(t *testing.T) {
	tests := []struct {
		name        string
		id          int
		newName     string
		mockSetup   func(mockRepo *repository.MockUserRepository, ctx context.Context, id int)
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Empty name",
			id:          1,
			newName:     "",
			mockSetup:   func(mockRepo *repository.MockUserRepository, ctx context.Context, id int) {},
			expectError: true,
			errorMsg:    "name cannot be empty",
		},
		{
			name:    "User not found/repo error",
			id:      1,
			newName: "New Name",
			mockSetup: func(mockRepo *repository.MockUserRepository, ctx context.Context, id int) {
				mockRepo.EXPECT().GetUserByID(ctx, id).Return(nil, errors.New("not found"))
			},
			expectError: true,
			errorMsg:    "not found",
		},
		{
			name:    "Successful update",
			id:      1,
			newName: "Updated Name",
			mockSetup: func(mockRepo *repository.MockUserRepository, ctx context.Context, id int) {
				user := &repository.User{ID: 1, Name: "Old Name"}
				mockRepo.EXPECT().GetUserByID(ctx, id).Return(user, nil)
				mockRepo.EXPECT().UpdateUser(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, u *repository.User) error {
					assert.Equal(t, "Updated Name", u.Name) // Verify that name was actually changed before update
					return nil
				})
			},
			expectError: false,
		},
		{
			name:    "UpdateUser Fails",
			id:      1,
			newName: "Updated Name",
			mockSetup: func(mockRepo *repository.MockUserRepository, ctx context.Context, id int) {
				user := &repository.User{ID: 1, Name: "Old Name"}
				mockRepo.EXPECT().GetUserByID(ctx, id).Return(user, nil)
				mockRepo.EXPECT().UpdateUser(ctx, gomock.Any()).Return(errors.New("update error"))
			},
			expectError: true,
			errorMsg:    "update error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRepo := repository.NewMockUserRepository(ctrl)
			userService := NewUserService(mockRepo)
			ctx := context.Background()

			tt.mockSetup(mockRepo, ctx, tt.id)
			err := userService.UpdateUserName(ctx, tt.id, tt.newName)

			if tt.expectError {
				require.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name        string
		id          int
		mockSetup   func(mockRepo *repository.MockUserRepository, ctx context.Context, id int)
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Attempt to delete admin",
			id:          1,
			mockSetup:   func(mockRepo *repository.MockUserRepository, ctx context.Context, id int) {},
			expectError: true,
			errorMsg:    "it is not allowed to delete admin user",
		},
		{
			name: "Successfull delete",
			id:   2,
			mockSetup: func(mockRepo *repository.MockUserRepository, ctx context.Context, id int) {
				mockRepo.EXPECT().DeleteUser(ctx, id).Return(nil)
			},
			expectError: false,
		},
		{
			name: "Repository Error",
			id:   2,
			mockSetup: func(mockRepo *repository.MockUserRepository, ctx context.Context, id int) {
				mockRepo.EXPECT().DeleteUser(ctx, id).Return(errors.New("db error"))
			},
			expectError: true,
			errorMsg:    "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRepo := repository.NewMockUserRepository(ctrl)
			userService := NewUserService(mockRepo)
			ctx := context.Background()

			tt.mockSetup(mockRepo, ctx, tt.id)
			err := userService.DeleteUser(ctx, tt.id)

			if tt.expectError {
				require.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
