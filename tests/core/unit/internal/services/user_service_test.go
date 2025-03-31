package unit

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/docker/go-connections/nat"
	_ "github.com/lib/pq" // cursor-- Added to register the Postgres SQL driver for wait.ForSQL
	"github.com/pageza/alchemorsel-v1/internal/db"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
	"github.com/pageza/alchemorsel-v1/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestMain sets up a PostgreSQL container for unit tests.
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "alchemorsel-test-*")
	if err != nil {
		fmt.Printf("Failed to create temp directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	req := testcontainers.ContainerRequest{
		Image:        "postgres:13", // Use a stable Postgres version matching production
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForSQL("5432/tcp", "postgres", func(host string, port nat.Port) string {
			return fmt.Sprintf("host=%s port=%s user=postgres password=testpass dbname=testdb sslmode=disable", host, port.Port())
		}).WithStartupTimeout(60 * time.Second),
	}

	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		fmt.Printf("Failed to start PostgreSQL container: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		_ = postgresC.Terminate(ctx)
	}()

	host, err := postgresC.Host(ctx)
	if err != nil {
		fmt.Printf("Failed to get container host: %v\n", err)
		os.Exit(1)
	}
	mappedPort, err := postgresC.MappedPort(ctx, "5432/tcp")
	if err != nil {
		fmt.Printf("Failed to get mapped port: %v\n", err)
		os.Exit(1)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=postgres password=testpass dbname=testdb sslmode=disable options='-c search_path=public,pg_catalog'", host, mappedPort.Port())

	db.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		os.Exit(1)
	}

	if err := db.DB.AutoMigrate(&models.User{}); err != nil {
		fmt.Printf("Failed to auto-migrate: %v\n", err)
		os.Exit(1)
	}

	if err := repositories.InitializeDB(dsn); err != nil {
		fmt.Printf("Failed to initialize repositories DB: %v\n", err)
		os.Exit(1)
	}

	// Run migrations for all models to create tables, including "recipes".
	if err := repositories.AutoMigrate(); err != nil {
		fmt.Printf("Failed to auto-migrate: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	os.Exit(code)
}

// MockUserRepository implements the UserRepository interface for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUser(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByResetPasswordToken(ctx context.Context, token string) (*models.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	service := services.NewUserService(mockRepo)

	tests := []struct {
		name    string
		user    *models.User
		wantErr bool
		setup   func()
	}{
		{
			name: "successful user creation",
			user: &models.User{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "Test1234!",
			},
			wantErr: false,
			setup: func() {
				mockRepo.On("GetUserByEmail", ctx, "test@example.com").Return(nil, nil)
				mockRepo.On("CreateUser", ctx, mock.AnythingOfType("*models.User")).Return(nil)
			},
		},
		{
			name: "duplicate email",
			user: &models.User{
				Name:     "Test User",
				Email:    "existing@example.com",
				Password: "Test1234!",
			},
			wantErr: true,
			setup: func() {
				mockRepo.On("GetUserByEmail", ctx, "existing@example.com").Return(&models.User{}, nil)
			},
		},
		{
			name: "invalid password",
			user: &models.User{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "weak",
			},
			wantErr: true,
			setup:   func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil
			tt.setup()

			err := service.CreateUser(ctx, tt.user)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, "Test1234!", tt.user.Password, "Password should be hashed")
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	service := services.NewUserService(mockRepo)

	expectedUser := &models.User{
		ID:    "123",
		Name:  "Test User",
		Email: "test@example.com",
	}

	mockRepo.On("GetUser", ctx, "123").Return(expectedUser, nil)
	mockRepo.On("GetUser", ctx, "456").Return(nil, fmt.Errorf("user not found"))

	t.Run("successful get user", func(t *testing.T) {
		user, err := service.GetUser(ctx, "123")
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
	})

	t.Run("user not found", func(t *testing.T) {
		user, err := service.GetUser(ctx, "456")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, "user not found", err.Error())
	})
}

func TestUpdateUser(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	service := services.NewUserService(mockRepo)

	user := &models.User{
		ID:       "123",
		Name:     "Updated Name",
		Email:    "updated@example.com",
		Password: "Test1234!",
	}

	mockRepo.On("UpdateUser", ctx, user).Return(nil)
	mockRepo.On("GetUserByEmail", ctx, "updated@example.com").Return(nil, nil)

	err := service.UpdateUser(ctx, "123", user)
	assert.NoError(t, err)
}

func TestDeleteUser(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	service := services.NewUserService(mockRepo)

	mockRepo.On("DeleteUser", ctx, "123").Return(nil)
	mockRepo.On("DeleteUser", ctx, "456").Return(fmt.Errorf("user not found"))

	t.Run("successful delete", func(t *testing.T) {
		err := service.DeleteUser(ctx, "123")
		assert.NoError(t, err)
	})

	t.Run("user not found", func(t *testing.T) {
		err := service.DeleteUser(ctx, "456")
		assert.Error(t, err)
		assert.Equal(t, "user not found", err.Error())
	})
}

func TestGetAllUsers(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	service := services.NewUserService(mockRepo)

	expectedUsers := []*models.User{
		{ID: "1", Name: "User 1", Email: "user1@example.com"},
		{ID: "2", Name: "User 2", Email: "user2@example.com"},
	}

	mockRepo.On("GetAllUsers", ctx).Return(expectedUsers, nil)

	users, err := service.GetAllUsers(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedUsers, users)
}

func TestResetPassword(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockUserRepository)
	service := services.NewUserService(mockRepo)

	expiry := time.Now().Add(24 * time.Hour)
	user := &models.User{
		ID:                   "123",
		ResetPasswordToken:   "valid-token",
		ResetPasswordExpires: &expiry,
	}

	mockRepo.On("GetUserByResetPasswordToken", ctx, "valid-token").Return(user, nil)
	mockRepo.On("UpdateUser", ctx, mock.AnythingOfType("*models.User")).Return(nil)
	mockRepo.On("GetUserByResetPasswordToken", ctx, "invalid-token").Return(nil, nil)

	t.Run("successful password reset", func(t *testing.T) {
		err := service.ResetPassword(ctx, "valid-token", "NewPassword123!")
		assert.NoError(t, err)
	})

	t.Run("invalid token", func(t *testing.T) {
		err := service.ResetPassword(ctx, "invalid-token", "NewPassword123!")
		assert.Error(t, err)
		assert.Equal(t, "invalid or expired reset token", err.Error())
	})
}

func TestUserService(t *testing.T) {
	mockRepo := new(MockUserRepository)
	_ = services.NewUserService(mockRepo)

	// TODO: Implement actual test cases using the mock repository
	// For example:
	// - Test user creation
	// - Test user retrieval
	// - Test user update
	// - Test user deletion
	// - Test password reset flow
}
