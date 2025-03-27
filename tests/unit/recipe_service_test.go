package unit

import (
	"context"
	"testing"
	"time"

	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/services"
	"github.com/stretchr/testify/assert"
)

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
	GetRecipeFunc    func(ctx context.Context, id string) (*models.Recipe, error)
	SaveRecipeFunc   func(ctx context.Context, recipe *models.Recipe) error
	ListRecipesFunc  func(ctx context.Context) ([]*models.Recipe, error)
	UpdateRecipeFunc func(ctx context.Context, recipe *models.Recipe) error
	DeleteRecipeFunc func(ctx context.Context, id string) error
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

func (m *MockRecipeRepository) ListRecipes(ctx context.Context) ([]*models.Recipe, error) {
	if m.ListRecipesFunc != nil {
		return m.ListRecipesFunc(ctx)
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

	// Create a new recipe with minimal fields.
	recipe := &models.Recipe{
		Title: "Test Recipe",
	}

	// Instantiate the service with the mock repository.
	service := services.NewRecipeService(mockRepo)

	err := service.SaveRecipe(context.Background(), recipe)
	assert.Nil(t, err, "Expected no error on saving recipe")
	assert.False(t, recipe.CreatedAt.IsZero(), "CreatedAt should be set")
	assert.False(t, recipe.UpdatedAt.IsZero(), "UpdatedAt should be set")
	assert.True(t, recipe.UpdatedAt.Equal(recipe.CreatedAt) || recipe.UpdatedAt.After(recipe.CreatedAt),
		"UpdatedAt should be equal to or after CreatedAt")
}

func TestSaveRecipeValidation(t *testing.T) {
	mockRepo := &MockRecipeRepository{}
	service := services.NewRecipeService(mockRepo)

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

	service := services.NewRecipeService(mockRepo)
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

	service := services.NewRecipeService(mockRepo)

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
	mockRecipes := []*models.Recipe{
		{ID: "1", Title: "Recipe 1"},
		{ID: "2", Title: "Recipe 2"},
	}

	mockRepo := &MockRecipeRepository{
		ListRecipesFunc: func(ctx context.Context) ([]*models.Recipe, error) {
			return mockRecipes, nil
		},
	}

	service := services.NewRecipeService(mockRepo)

	recipes, err := service.ListRecipes(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, mockRecipes, recipes)
}

func TestListRecipesError(t *testing.T) {
	mockRepo := &MockRecipeRepository{
		ListRecipesFunc: func(ctx context.Context) ([]*models.Recipe, error) {
			return nil, assert.AnError
		},
	}

	service := services.NewRecipeService(mockRepo)

	recipes, err := service.ListRecipes(context.Background())
	assert.Error(t, err)
	assert.Nil(t, recipes)
}

func TestUpdateRecipe(t *testing.T) {
	mockRepo := &MockRecipeRepository{
		UpdateRecipeFunc: func(ctx context.Context, recipe *models.Recipe) error {
			return nil
		},
	}

	service := services.NewRecipeService(mockRepo)

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

	service := services.NewRecipeService(mockRepo)

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
