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
	"github.com/stretchr/testify/mock"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestMain sets up a PostgreSQL container for unit tests.
func TestMain(m *testing.M) {
	ctx := context.Background()

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

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	user := &models.User{
		Name:  "Test User",
		Email: "test@example.com",
		// Updated password: at least 8 characters with one digit, one uppercase, one lowercase, and one special character.
		Password: "Test1234!",
	}

	mockRepo := new(MockUserRepository)
	service := services.NewUserService(mockRepo)

	mockRepo.On("GetUserByEmail", ctx, user.Email).Return(nil, nil)
	mockRepo.On("CreateUser", ctx, mock.AnythingOfType("*models.User")).Return(nil)

	err := service.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("User creation failed: %v", err)
	}
	// Verify that the plain text password is replaced with a hashed one.
	if user.Password == "Test1234!" {
		t.Errorf("Expected password to be hashed, but it remains in plain text")
	}
}

// Replace db.DB usage with mock repositories
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
