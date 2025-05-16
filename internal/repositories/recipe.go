package repositories

import "encoding/json"

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
