package repositories

import (
	"encoding/json"

	"github.com/pageza/alchemorsel-v1/internal/models"
)

// Recipe represents a recipe in the repository layer
type Recipe struct {
	ID               string        `json:"id"`
	Title            string        `json:"title"`
	Description      string        `json:"description"`
	Servings         int           `json:"servings"`
	PrepTimeMinutes  int           `json:"prep_time_minutes"`
	CookTimeMinutes  int           `json:"cook_time_minutes"`
	TotalTimeMinutes int           `json:"total_time_minutes"`
	Ingredients      []Ingredient  `json:"ingredients"`
	Instructions     []Instruction `json:"instructions"`
	Nutrition        Nutrition     `json:"nutrition"`
	Tags             []string      `json:"tags"`
	Difficulty       string        `json:"difficulty"`
	UserID           string        `json:"user_id"`
}

// ToModel converts a repository Recipe to a model Recipe
func (r *Recipe) ToModel() *models.Recipe {
	return &models.Recipe{
		ID:               r.ID,
		Title:            r.Title,
		Description:      r.Description,
		Servings:         r.Servings,
		PrepTimeMinutes:  r.PrepTimeMinutes,
		CookTimeMinutes:  r.CookTimeMinutes,
		TotalTimeMinutes: r.TotalTimeMinutes,
		Ingredients:      convertToModelIngredients(r.Ingredients),
		Instructions:     convertToModelInstructions(r.Instructions),
		Nutrition:        convertToModelNutrition(r.Nutrition),
		Tags:             r.Tags,
		Difficulty:       r.Difficulty,
		UserID:           r.UserID,
	}
}

// FromModel converts a model Recipe to a repository Recipe
func FromModel(m *models.Recipe) *Recipe {
	return &Recipe{
		ID:               m.ID,
		Title:            m.Title,
		Description:      m.Description,
		Servings:         m.Servings,
		PrepTimeMinutes:  m.PrepTimeMinutes,
		CookTimeMinutes:  m.CookTimeMinutes,
		TotalTimeMinutes: m.TotalTimeMinutes,
		Ingredients:      convertToRepoIngredients(m.Ingredients),
		Instructions:     convertToRepoInstructions(m.Instructions),
		Nutrition:        convertToRepoNutrition(m.Nutrition),
		Tags:             m.Tags,
		Difficulty:       m.Difficulty,
		UserID:           m.UserID,
	}
}

type Ingredient struct {
	Item   string      `json:"item"`
	Amount json.Number `json:"amount"`
	Unit   string      `json:"unit"`
}

type Instruction struct {
	Step        int    `json:"step"`
	Description string `json:"description"`
}

type Nutrition struct {
	Calories int    `json:"calories"`
	Protein  string `json:"protein"`
	Carbs    string `json:"carbs"`
	Fat      string `json:"fat"`
}

// Helper functions to convert between repository and model types
func convertToModelIngredients(repoIngredients []Ingredient) []models.Ingredient {
	modelIngredients := make([]models.Ingredient, len(repoIngredients))
	for i, ing := range repoIngredients {
		modelIngredients[i] = models.Ingredient{
			Item:   ing.Item,
			Amount: ing.Amount,
			Unit:   ing.Unit,
		}
	}
	return modelIngredients
}

func convertToModelInstructions(repoInstructions []Instruction) []models.Instruction {
	modelInstructions := make([]models.Instruction, len(repoInstructions))
	for i, inst := range repoInstructions {
		modelInstructions[i] = models.Instruction{
			Step:        inst.Step,
			Description: inst.Description,
		}
	}
	return modelInstructions
}

func convertToModelNutrition(repoNutrition Nutrition) models.Nutrition {
	return models.Nutrition{
		Calories: repoNutrition.Calories,
		Protein:  repoNutrition.Protein,
		Carbs:    repoNutrition.Carbs,
		Fat:      repoNutrition.Fat,
	}
}

func convertToRepoIngredients(modelIngredients []models.Ingredient) []Ingredient {
	repoIngredients := make([]Ingredient, len(modelIngredients))
	for i, ing := range modelIngredients {
		repoIngredients[i] = Ingredient{
			Item:   ing.Item,
			Amount: ing.Amount,
			Unit:   ing.Unit,
		}
	}
	return repoIngredients
}

func convertToRepoInstructions(modelInstructions []models.Instruction) []Instruction {
	repoInstructions := make([]Instruction, len(modelInstructions))
	for i, inst := range modelInstructions {
		repoInstructions[i] = Instruction{
			Step:        inst.Step,
			Description: inst.Description,
		}
	}
	return repoInstructions
}

func convertToRepoNutrition(modelNutrition models.Nutrition) Nutrition {
	return Nutrition{
		Calories: modelNutrition.Calories,
		Protein:  modelNutrition.Protein,
		Carbs:    modelNutrition.Carbs,
		Fat:      modelNutrition.Fat,
	}
}
