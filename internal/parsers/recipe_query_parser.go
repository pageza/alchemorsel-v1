package parsers

import (
	"errors"
	"strings"

	"github.com/jdkato/prose/v2"
)

// ParsedQuery represents the structured output from parsing a user's freeform recipe query.
type ParsedQuery struct {
	Cuisine             string   `json:"cuisine"`
	DietaryRestrictions string   `json:"dietary_restrictions"`
	Ingredients         []string `json:"ingredients"`
	Exclusions          []string `json:"exclusions"`

	// Additional optional filters for more detailed queries
	Timing             int    `json:"timing,omitempty"`               // Total time in minutes (prep + cooking)
	Servings           int    `json:"servings,omitempty"`             // Desired number of servings
	Difficulty         string `json:"difficulty,omitempty"`           // e.g., "easy", "medium", "hard"
	CaloriesPerServing int    `json:"calories_per_serving,omitempty"` // Maximum calories per serving
	ServingSize        string `json:"serving_size,omitempty"`         // e.g., "small", "medium", "large"
}

// ParseRecipeQuery parses the user's freeform query into a structured ParsedQuery using the prose NLP library.
// This implementation uses tokenization and basic part-of-speech tagging to extract information,
// including handling exclusions when a user specifies they don't want an ingredient (e.g., "no onions").
func ParseRecipeQuery(query string) (*ParsedQuery, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("empty query")
	}

	doc, err := prose.NewDocument(query)
	if err != nil {
		return nil, err
	}

	// Define known cuisines and dietary restrictions
	knownCuisines := []string{"mexican", "italian", "asian", "french", "chinese", "indian"}
	knownDietary := []string{"vegan", "vegetarian", "paleo", "gluten-free", "ketogenic"}

	cuisine := "unknown"
	dietary := "none"
	ingredients := []string{}
	exclusions := []string{}
	tokens := doc.Tokens()

	for i, tok := range tokens {
		lowerToken := strings.ToLower(tok.Text)

		// Check if the token matches any known cuisines
		for _, c := range knownCuisines {
			if lowerToken == c {
				cuisine = c
			}
		}

		// Check if the token matches any known dietary restrictions
		for _, d := range knownDietary {
			if lowerToken == d {
				dietary = d
			}
		}

		// Check if token is a noun (ingredient candidate)
		if strings.HasPrefix(tok.Tag, "NN") {
			// If the previous token is "no" or "without", add to exclusions
			if i > 0 {
				prevToken := strings.ToLower(tokens[i-1].Text)
				if prevToken == "no" || prevToken == "without" {
					exclusions = append(exclusions, lowerToken)
					continue
				}
			}
			// Otherwise, add to ingredients (avoid duplicate insertion if already captured as cuisine/dietary)
			if lowerToken != cuisine && lowerToken != dietary {
				ingredients = append(ingredients, lowerToken)
			}
		}
	}

	pq := &ParsedQuery{
		Cuisine:             cuisine,
		DietaryRestrictions: dietary,
		Ingredients:         ingredients,
		Exclusions:          exclusions,
		Timing:              0,
		Servings:            0,
		Difficulty:          "",
		CaloriesPerServing:  0,
		ServingSize:         "",
	}

	// TODO: Implement enhanced NLP parsing here to extract tokens for timing, servings, difficulty, etc.
	// For now, if the user does not explicitly provide these details, the values remain default.

	return pq, nil
}
