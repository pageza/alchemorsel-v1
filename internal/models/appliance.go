package models

// Appliance represents a cooking appliance required by a recipe (e.g., frying pan, oven, blender).
type Appliance struct {
	ID   string `json:"id" gorm:"type:uuid;primaryKey"`
	Name string `json:"name" gorm:"unique;not null"`
}
