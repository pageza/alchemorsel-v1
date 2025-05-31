package unit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock services
type MockCuisineService struct{}

func (m *MockCuisineService) GetByID(ctx context.Context, id string) (*models.Cuisine, error) {
	return &models.Cuisine{ID: id, Name: "test"}, nil
}
func (m *MockCuisineService) GetByName(ctx context.Context, name string) (*models.Cuisine, error) {
	return &models.Cuisine{ID: "test-id", Name: name}, nil
}
func (m *MockCuisineService) Create(ctx context.Context, cuisine *models.Cuisine) error {
	return nil
}
func (m *MockCuisineService) List(ctx context.Context) ([]*models.Cuisine, error) {
	return []*models.Cuisine{}, nil
}
func (m *MockCuisineService) Delete(ctx context.Context, id string) error {
	return nil
}
func (m *MockCuisineService) GetOrCreate(ctx context.Context, name string) (*models.Cuisine, error) {
	return &models.Cuisine{ID: "test-id", Name: name}, nil
}

type MockDietService struct{}

func (m *MockDietService) GetByID(ctx context.Context, id string) (*models.Diet, error) {
	return &models.Diet{ID: id, Name: "test"}, nil
}
func (m *MockDietService) GetByName(ctx context.Context, name string) (*models.Diet, error) {
	return &models.Diet{ID: "test-id", Name: name}, nil
}
func (m *MockDietService) Create(ctx context.Context, diet *models.Diet) error {
	return nil
}
func (m *MockDietService) List(ctx context.Context) ([]*models.Diet, error) {
	return []*models.Diet{}, nil
}
func (m *MockDietService) Delete(ctx context.Context, id string) error {
	return nil
}
func (m *MockDietService) GetOrCreate(ctx context.Context, name string) (*models.Diet, error) {
	return &models.Diet{ID: "test-id", Name: name}, nil
}

type MockApplianceService struct{}

func (m *MockApplianceService) GetByID(ctx context.Context, id string) (*models.Appliance, error) {
	return &models.Appliance{ID: id, Name: "test"}, nil
}
func (m *MockApplianceService) GetByName(ctx context.Context, name string) (*models.Appliance, error) {
	return &models.Appliance{ID: "test-id", Name: name}, nil
}
func (m *MockApplianceService) Create(ctx context.Context, appliance *models.Appliance) error {
	return nil
}
func (m *MockApplianceService) List(ctx context.Context) ([]*models.Appliance, error) {
	return []*models.Appliance{}, nil
}
func (m *MockApplianceService) Delete(ctx context.Context, id string) error {
	return nil
}
func (m *MockApplianceService) GetOrCreate(ctx context.Context, name string) (*models.Appliance, error) {
	return &models.Appliance{ID: "test-id", Name: name}, nil
}

type MockTagService struct{}

func (m *MockTagService) GetByID(ctx context.Context, id string) (*models.Tag, error) {
	return &models.Tag{ID: id, Name: "test"}, nil
}
func (m *MockTagService) GetByName(ctx context.Context, name string) (*models.Tag, error) {
	return &models.Tag{ID: "test-id", Name: name}, nil
}
func (m *MockTagService) Create(ctx context.Context, tag *models.Tag) error {
	return nil
}
func (m *MockTagService) List(ctx context.Context) ([]*models.Tag, error) {
	return []*models.Tag{}, nil
}
func (m *MockTagService) Delete(ctx context.Context, id string) error {
	return nil
}
func (m *MockTagService) GetOrCreate(ctx context.Context, name string) (*models.Tag, error) {
	return &models.Tag{ID: "test-id", Name: name}, nil
}

// func TestListRecipes(t *testing.T) {
// 	recipes, err := services.ListRecipes()
// 	if err == nil {
// 		t.Error("Expected error for unimplemented ListRecipes, got nil")
// 	}
// 	if recipes != nil {
// 		t.Error("Expected nil recipes for unimplemented ListRecipes")
// 	}
// }

// MockRecipeRepository is a mock implementation of RecipeRepository for testing.
type MockRecipeRepository struct {
	GetRecipeFunc        func(ctx context.Context, id string) (*models.Recipe, error)
	SaveRecipeFunc       func(ctx context.Context, recipe *models.Recipe) error
	ListRecipesFunc      func(ctx context.Context, page, limit int, sort, order string) ([]models.Recipe, error)
	UpdateRecipeFunc     func(ctx context.Context, recipe *models.Recipe) error
	DeleteRecipeFunc     func(ctx context.Context, id string) error
	SearchRecipesFunc    func(ctx context.Context, query string, tags []string, difficulty string) ([]models.Recipe, error)
	RateRecipeFunc       func(ctx context.Context, recipeID string, rating float64) error
	GetRecipeRatingsFunc func(ctx context.Context, recipeID string) ([]float64, error)
	ResolveRecipeFunc    func(ctx context.Context, query string, attributes map[string]interface{}) (*models.Recipe, []*models.Recipe, error)
}

func (m *MockRecipeRepository) GetRecipe(ctx context.Context, id string) (*models.Recipe, error) {
	if m.GetRecipeFunc != nil {
		return m.GetRecipeFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockRecipeRepository) SaveRecipe(ctx context.Context, recipe *models.Recipe) error {
	if m.SaveRecipeFunc != nil {
		return m.SaveRecipeFunc(ctx, recipe)
	}
	return nil
}

func (m *MockRecipeRepository) ListRecipes(ctx context.Context, page, limit int, sort, order string) ([]models.Recipe, error) {
	if m.ListRecipesFunc != nil {
		return m.ListRecipesFunc(ctx, page, limit, sort, order)
	}
	return nil, nil
}

func (m *MockRecipeRepository) UpdateRecipe(ctx context.Context, recipe *models.Recipe) error {
	if m.UpdateRecipeFunc != nil {
		return m.UpdateRecipeFunc(ctx, recipe)
	}
	return nil
}

func (m *MockRecipeRepository) DeleteRecipe(ctx context.Context, id string) error {
	if m.DeleteRecipeFunc != nil {
		return m.DeleteRecipeFunc(ctx, id)
	}
	return nil
}

func (m *MockRecipeRepository) SearchRecipes(ctx context.Context, query string, tags []string, difficulty string) ([]models.Recipe, error) {
	if m.SearchRecipesFunc != nil {
		return m.SearchRecipesFunc(ctx, query, tags, difficulty)
	}
	return nil, nil
}

func (m *MockRecipeRepository) RateRecipe(ctx context.Context, recipeID string, rating float64) error {
	if m.RateRecipeFunc != nil {
		return m.RateRecipeFunc(ctx, recipeID, rating)
	}
	return nil
}

func (m *MockRecipeRepository) GetRecipeRatings(ctx context.Context, recipeID string) ([]float64, error) {
	if m.GetRecipeRatingsFunc != nil {
		return m.GetRecipeRatingsFunc(ctx, recipeID)
	}
	return nil, nil
}

func (m *MockRecipeRepository) ResolveRecipe(ctx context.Context, query string, attributes map[string]interface{}) (*models.Recipe, []*models.Recipe, error) {
	if m.ResolveRecipeFunc != nil {
		return m.ResolveRecipeFunc(ctx, query, attributes)
	}
	return nil, nil, nil
}

func TestSaveRecipeSuccess(t *testing.T) {
	// Create a mock repository that simulates a successful save.
	mockRepo := &MockRecipeRepository{
		SaveRecipeFunc: func(ctx context.Context, recipe *models.Recipe) error {
			// Set timestamps if they're not already set
			now := time.Now()
			if recipe.CreatedAt.IsZero() {
				recipe.CreatedAt = now
			}
			if recipe.UpdatedAt.IsZero() {
				recipe.UpdatedAt = now
			}
			return nil
		},
	}

	// Create mock services
	mockCuisineService := &MockCuisineService{}
	mockDietService := &MockDietService{}
	mockApplianceService := &MockApplianceService{}
	mockTagService := &MockTagService{}

	// Create a new recipe with minimal fields.
	recipe := &models.Recipe{
		Title: "Test Recipe",
	}

	// Instantiate the service with the mock repository and services.
	service := services.NewRecipeService(mockRepo, mockCuisineService, mockDietService, mockApplianceService, mockTagService)

	err := service.SaveRecipe(context.Background(), recipe)
	assert.Nil(t, err, "Expected no error on saving recipe")
	assert.False(t, recipe.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.False(t, recipe.UpdatedAt.IsZero(), "UpdatedAt should be set")
	assert.True(t, recipe.UpdatedAt.Equal(recipe.CreatedAt) || recipe.UpdatedAt.After(recipe.CreatedAt),
		"UpdatedAt should be equal to or after CreatedAt")
}

func TestSaveRecipeValidation(t *testing.T) {
	mockRepo := &MockRecipeRepository{}
	mockCuisineService := &MockCuisineService{}
	mockDietService := &MockDietService{}
	mockApplianceService := &MockApplianceService{}
	mockTagService := &MockTagService{}

	service := services.NewRecipeService(mockRepo, mockCuisineService, mockDietService, mockApplianceService, mockTagService)

	tests := []struct {
		name    string
		recipe  *models.Recipe
		wantErr bool
	}{
		{
			name:    "nil recipe",
			recipe:  nil,
			wantErr: true,
		},
		{
			name:    "empty title",
			recipe:  &models.Recipe{},
			wantErr: true,
		},
		{
			name: "valid recipe",
			recipe: &models.Recipe{
				Title: "Valid Recipe",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SaveRecipe(context.Background(), tt.recipe)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSaveRecipeDBError(t *testing.T) {
	mockRepo := &MockRecipeRepository{
		SaveRecipeFunc: func(ctx context.Context, recipe *models.Recipe) error {
			return assert.AnError
		},
	}

	mockCuisineService := &MockCuisineService{}
	mockDietService := &MockDietService{}
	mockApplianceService := &MockApplianceService{}
	mockTagService := &MockTagService{}

	service := services.NewRecipeService(mockRepo, mockCuisineService, mockDietService, mockApplianceService, mockTagService)
	recipe := &models.Recipe{
		Title: "Test Recipe",
	}

	err := service.SaveRecipe(context.Background(), recipe)
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

func TestGetRecipe(t *testing.T) {
	mockRepo := &MockRecipeRepository{
		GetRecipeFunc: func(ctx context.Context, id string) (*models.Recipe, error) {
			if id == "valid-id" {
				return &models.Recipe{
					ID:    "valid-id",
					Title: "Test Recipe",
				}, nil
			}
			return nil, assert.AnError
		},
	}

	mockCuisineService := &MockCuisineService{}
	mockDietService := &MockDietService{}
	mockApplianceService := &MockApplianceService{}
	mockTagService := &MockTagService{}

	service := services.NewRecipeService(mockRepo, mockCuisineService, mockDietService, mockApplianceService, mockTagService)

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "valid id",
			id:      "valid-id",
			wantErr: false,
		},
		{
			name:    "invalid id",
			id:      "invalid-id",
			wantErr: true,
		},
		{
			name:    "empty id",
			id:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipe, err := service.GetRecipe(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, recipe)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, recipe)
				assert.Equal(t, tt.id, recipe.ID)
			}
		})
	}
}

func TestListRecipes(t *testing.T) {
	mockRecipes := []models.Recipe{
		{
			ID:    "1",
			Title: "Test Recipe 1",
		},
		{
			ID:    "2",
			Title: "Test Recipe 2",
		},
	}

	mockRepo := &MockRecipeRepository{
		ListRecipesFunc: func(ctx context.Context, page, limit int, sort, order string) ([]models.Recipe, error) {
			return mockRecipes, nil
		},
	}

	mockCuisineService := &MockCuisineService{}
	mockDietService := &MockDietService{}
	mockApplianceService := &MockApplianceService{}
	mockTagService := &MockTagService{}

	service := services.NewRecipeService(mockRepo, mockCuisineService, mockDietService, mockApplianceService, mockTagService)

	recipes, err := service.ListRecipes(context.Background(), 1, 10, "created_at", "desc")
	assert.NoError(t, err)
	assert.Equal(t, mockRecipes, recipes)
}

func TestListRecipesError(t *testing.T) {
	mockRepo := &MockRecipeRepository{
		ListRecipesFunc: func(ctx context.Context, page, limit int, sort, order string) ([]models.Recipe, error) {
			return nil, assert.AnError
		},
	}

	mockCuisineService := &MockCuisineService{}
	mockDietService := &MockDietService{}
	mockApplianceService := &MockApplianceService{}
	mockTagService := &MockTagService{}

	service := services.NewRecipeService(mockRepo, mockCuisineService, mockDietService, mockApplianceService, mockTagService)

	recipes, err := service.ListRecipes(context.Background(), 1, 10, "created_at", "desc")
	assert.Error(t, err)
	assert.Nil(t, recipes)
}

func TestUpdateRecipe(t *testing.T) {
	mockRepo := &MockRecipeRepository{
		UpdateRecipeFunc: func(ctx context.Context, recipe *models.Recipe) error {
			return nil
		},
	}

	mockCuisineService := &MockCuisineService{}
	mockDietService := &MockDietService{}
	mockApplianceService := &MockApplianceService{}
	mockTagService := &MockTagService{}

	service := services.NewRecipeService(mockRepo, mockCuisineService, mockDietService, mockApplianceService, mockTagService)

	tests := []struct {
		name    string
		recipe  *models.Recipe
		wantErr bool
	}{
		{
			name: "valid update",
			recipe: &models.Recipe{
				ID:    "1",
				Title: "Updated Recipe",
			},
			wantErr: false,
		},
		{
			name:    "nil recipe",
			recipe:  nil,
			wantErr: true,
		},
		{
			name: "empty title",
			recipe: &models.Recipe{
				ID: "1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateRecipe(context.Background(), tt.recipe)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteRecipe(t *testing.T) {
	mockRepo := &MockRecipeRepository{
		DeleteRecipeFunc: func(ctx context.Context, id string) error {
			if id == "valid-id" {
				return nil
			}
			return assert.AnError
		},
	}

	mockCuisineService := &MockCuisineService{}
	mockDietService := &MockDietService{}
	mockApplianceService := &MockApplianceService{}
	mockTagService := &MockTagService{}

	service := services.NewRecipeService(mockRepo, mockCuisineService, mockDietService, mockApplianceService, mockTagService)

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "valid id",
			id:      "valid-id",
			wantErr: false,
		},
		{
			name:    "invalid id",
			id:      "invalid-id",
			wantErr: true,
		},
		{
			name:    "empty id",
			id:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeleteRecipe(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRecipeService_EdgeCases(t *testing.T) {
	mockRepo := new(MockRecipeRepository)
	service := services.NewRecipeService(mockRepo)

	t.Run("SaveRecipe_NilRecipe", func(t *testing.T) {
		err := service.SaveRecipe(context.Background(), nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "recipe cannot be nil")
	})

	t.Run("SaveRecipe_EmptyTitle", func(t *testing.T) {
		recipe := &models.Recipe{
			Title: "",
			Instructions: []models.Instruction{
				{StepNumber: 1, Description: "Test instruction"},
			},
		}
		
		err := service.SaveRecipe(context.Background(), recipe)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "title is required")
	})

	t.Run("SaveRecipe_EmptyInstructions", func(t *testing.T) {
		recipe := &models.Recipe{
			Title:        "Test Recipe",
			Instructions: []models.Instruction{},
		}
		
		err := service.SaveRecipe(context.Background(), recipe)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "instructions are required")
	})

	t.Run("GetRecipe_InvalidUUID", func(t *testing.T) {
		_, err := service.GetRecipe(context.Background(), "invalid-uuid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid UUID")
	})

	t.Run("UpdateRecipe_NonExistentRecipe", func(t *testing.T) {
		recipe := &models.Recipe{
			ID:    "550e8400-e29b-41d4-a716-446655440000",
			Title: "Updated Recipe",
			Instructions: []models.Instruction{
				{StepNumber: 1, Description: "Updated instruction"},
			},
		}

		mockRepo.On("GetRecipe", mock.Anything, recipe.ID).Return(nil, fmt.Errorf("recipe not found"))

		err := service.UpdateRecipe(context.Background(), recipe)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "recipe not found")
	})

	t.Run("DeleteRecipe_InvalidUUID", func(t *testing.T) {
		err := service.DeleteRecipe(context.Background(), "invalid-uuid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid UUID")
	})

	t.Run("SearchRecipes_EmptyQuery", func(t *testing.T) {
		recipes, err := service.SearchRecipes(context.Background(), "")
		assert.NoError(t, err)
		assert.Empty(t, recipes)
	})

	t.Run("SaveRecipe_LargeRecipeData", func(t *testing.T) {
		largeInstructions := make([]models.Instruction, 1000)
		for i := 0; i < 1000; i++ {
			largeInstructions[i] = models.Instruction{
				StepNumber:  i + 1,
				Description: fmt.Sprintf("Very long instruction text %d", i),
			}
		}

		recipe := &models.Recipe{
			Title:        "Large Recipe",
			Instructions: largeInstructions,
		}

		mockRepo.On("SaveRecipe", mock.Anything, recipe).Return(nil)

		err := service.SaveRecipe(context.Background(), recipe)
		assert.NoError(t, err)
	})
}
