package main

import (
	"fmt"
	"log"

	"github.com/pageza/alchemorsel-v1/internal/parsers"
)

func main() {
	query := "I want a Mexican vegan dish with tomatoes and without onions"
	parsed, err := parsers.ParseRecipeQuery(query)
	if err != nil {
		log.Fatalf("Error parsing query: %v", err)
	}
	fmt.Println("Parsed Query Result:")
	fmt.Printf("  Cuisine: %s\n", parsed.Cuisine)
	fmt.Printf("  Dietary Restrictions: %s\n", parsed.DietaryRestrictions)
	fmt.Printf("  Ingredients: %v\n", parsed.Ingredients)
	fmt.Printf("  Exclusions: %v\n", parsed.Exclusions)
}
