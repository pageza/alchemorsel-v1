# API Contract Reference for MVP

## 1. Overview

This document serves as a central reference for standardizing our API contract as part of our MVP. It consolidates our review of current data models, outlines the standardized response formats, and details the action plan to ensure consistency across core endpoints.

## 2. Review of Current Data Models

### Recipe Model

- **Location:** `internal/models/recipe.go`
- **Key Fields:**
  - `ID`: string (UUID)
  - `Title`: string (required)
  - `Description`: string
  - `Ingredients`: JSON (expected to be an array of objects with `name`, `amount`, and `unit`)
  - `Steps`: JSON (expected to be an array of objects with `order` and `description`)
  - *Other fields:* Nutritional info, allergy disclaimer, related entities (cuisines, diets, appliances, tags), images, difficulty, prep/cook time, servings, rating, timestamps, approved status, and embedding data

> **Observation:** Some tests previously expected `steps` as a simple string. Our current (and intended) design uses structured JSON arrays for both ingredients and steps.

### User Model

- **Location:** `internal/models/user.go`
- **Key Fields:**
  - `ID`: string (UUID or serial as defined)
  - `Email`: string (unique, required)
  - `PasswordHash`: string
  - Additional profile fields (name, created/updated timestamps, etc.)

> **Observation:** There is some discrepancy between what some e2e tests expect (additional fields like bio, preferences) and the core model. We need to decide on the minimal set for MVP and document it.

### Other Models (Cuisine, Diet, Appliance, Tag, etc.)

- These models are used primarily to maintain many-to-many relationships with recipes. They typically include an `ID` and a `Name` field. Their JSON representations will be consistent across endpoints.

## 3. Standardized API Contract Specification

For our MVP, we propose the following standardized response formats:

### Recipe GET Endpoint Response Example

```json
{
  "id": "<UUID>",
  "title": "Spaghetti Bolognese",
  "description": "A hearty Italian pasta dish.",
  "ingredients": [
    { "name": "Spaghetti", "amount": "200", "unit": "grams" },
    { "name": "Ground Beef", "amount": "300", "unit": "grams" }
  ],
  "steps": [
    { "order": 1, "description": "Boil water and cook spaghetti." },
    { "order": 2, "description": "Cook beef with tomato sauce." }
  ],
  "nutritional_info": "Calories: 500...",
  "allergy_disclaimer": "Contains gluten",
  "cuisines": [ /* Array of related cuisines */ ],
  "diets": [ /* Array of related diets */ ],
  "appliances": [ /* Array of related appliances */ ],
  "tags": [ /* Array of tags */ ],
  "images": [ /* Array of image URLs or data */ ],
  "difficulty": "Medium",
  "prep_time": 15,
  "cooking_time": 30,
  "servings": 4,
  "average_rating": 4.5,
  "rating_count": 100,
  "created_at": "2023-10-04T12:00:00Z",
  "updated_at": "2023-10-04T12:00:00Z",
  "approved": true,
  "embedding": [0.1234, 0.5678, 0.9101]
}
```

### User GET Endpoint Response Example

For MVP, the user endpoint may return a simplified version:

```json
{
  "id": "<UUID>",
  "email": "user@example.com",
  "name": "John Doe",
  "created_at": "2023-10-04T12:00:00Z",
  "updated_at": "2023-10-04T12:00:00Z"
}
```

## 4. Action Plan for Standardizing the API Contract (Step 1 of Game Plan)

1. **Formalize the API Contract**:
   - Draft the specification as shown above for the Recipe and User endpoints (and extend as needed for other entities).
   - Confirm with the team whether to extend the User model or adjust test expectations.

2. **Update Implementation & Tests**:
   - Ensure that handler methods for recipes and users serialize responses according to the above contract.
   - Audit tests to update any expectations on data structures—especially ensuring that `ingredients` and `steps` are treated as arrays.
   - Update mocks in unit tests to align with the standardized contract.

3. **Document & Version the API Contract**:
   - Host this document (e.g., within our repository's docs or README) as the definitive API contract.
   - Use versioning to track future changes so that test suites across unit, integration, and e2e phases can be updated in lockstep.

## 5. Next Steps

- **Implementation:** Update core business logic and test expectations to adhere to this contract.
- **Testing:** After these updates, run core tests (unit, integration, and e2e) and verify that they align with this contract.
- **Peripheral Tests:** Temporarily disable peripheral tests (using `t.Skip` or commenting out) until core functionality is fully stabilized.

---

This document, along with the Test Audit and Game Plan (in `TEST_AUDIT_AND_GAMEPLAN.md`), will serve as our main references for ensuring consistency across all testing phases during our MVP development. 