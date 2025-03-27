package api

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewAPIDocManager(t *testing.T) {
	version := "1.0.0"
	manager := NewAPIDocManager(version)

	assert.NotNil(t, manager)
	assert.Equal(t, version, manager.version)
	assert.NotNil(t, manager.doc)

	// Check basic info
	assert.Equal(t, "Alchemorsel API", manager.doc.Info.Title)
	assert.Equal(t, "API for the Alchemorsel application", manager.doc.Info.Description)
	assert.Equal(t, version, manager.doc.Info.Version)
	assert.Equal(t, "Alchemorsel Team", manager.doc.Info.Contact.Name)
	assert.Equal(t, "support@alchemorsel.com", manager.doc.Info.Contact.Email)
	assert.Equal(t, "MIT", manager.doc.Info.License.Name)
	assert.Equal(t, "https://opensource.org/licenses/MIT", manager.doc.Info.License.URL)

	// Check servers
	assert.Len(t, manager.doc.Servers, 1)
	assert.Equal(t, "/api/v"+version, manager.doc.Servers[0].URL)
	assert.Equal(t, "API v"+version, manager.doc.Servers[0].Description)

	// Check components
	assert.NotNil(t, manager.doc.Components.Schemas)
	assert.NotNil(t, manager.doc.Components.SecuritySchemes)
}

func TestAddPath(t *testing.T) {
	manager := NewAPIDocManager("1.0.0")

	// Create a test path item
	pathItem := PathItem{
		Get: &Operation{
			Summary:     "Get user",
			Description: "Retrieve user information",
			Tags:        []string{"users"},
			Responses: map[string]Response{
				"200": {
					Description: "Success",
					Content: map[string]MediaType{
						"application/json": {
							Schema: Schema{
								Ref: "#/components/schemas/User",
							},
						},
					},
				},
			},
		},
	}

	// Add the path
	path := "/users/{id}"
	manager.AddPath(path, pathItem)

	// Verify the path was added
	assert.Contains(t, manager.doc.Paths, path)
	assert.Equal(t, pathItem, manager.doc.Paths[path])
}

func TestAddSchema(t *testing.T) {
	manager := NewAPIDocManager("1.0.0")

	// Create a test schema
	schema := Schema{
		Type: "object",
		Properties: map[string]Schema{
			"id": {
				Type:   "string",
				Format: "uuid",
			},
			"name": {
				Type: "string",
			},
		},
		Required: []string{"id", "name"},
	}

	// Add the schema
	schemaName := "User"
	manager.AddSchema(schemaName, schema)

	// Verify the schema was added
	assert.Contains(t, manager.doc.Components.Schemas, schemaName)
	assert.Equal(t, schema, manager.doc.Components.Schemas[schemaName])
}

func TestAddSecurityScheme(t *testing.T) {
	manager := NewAPIDocManager("1.0.0")

	// Create a test security scheme
	scheme := SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Description:  "JWT token authentication",
	}

	// Add the security scheme
	schemeName := "bearerAuth"
	manager.AddSecurityScheme(schemeName, scheme)

	// Verify the security scheme was added
	assert.Contains(t, manager.doc.Components.SecuritySchemes, schemeName)
	assert.Equal(t, scheme, manager.doc.Components.SecuritySchemes[schemeName])
}

func TestGenerateMarkdown(t *testing.T) {
	manager := NewAPIDocManager("1.0.0")

	// Add a test path
	pathItem := PathItem{
		Get: &Operation{
			Summary:     "Get user",
			Description: "Retrieve user information",
			Tags:        []string{"users"},
		},
		Post: &Operation{
			Summary:     "Create user",
			Description: "Create a new user",
			Tags:        []string{"users"},
		},
	}
	manager.AddPath("/users", pathItem)

	// Generate markdown
	markdown, err := manager.GenerateMarkdown()
	assert.NoError(t, err)
	assert.NotEmpty(t, markdown)

	// Verify markdown content
	assert.Contains(t, markdown, "# Alchemorsel API")
	assert.Contains(t, markdown, "Version: 1.0.0")
	assert.Contains(t, markdown, "## Endpoints")
	assert.Contains(t, markdown, "### /users")
	assert.Contains(t, markdown, "#### GET")
	assert.Contains(t, markdown, "#### POST")
}

func TestRegisterSwagger(t *testing.T) {
	manager := NewAPIDocManager("1.0.0")
	router := gin.New()

	// Register Swagger routes
	manager.RegisterSwagger(router)

	// Verify routes are registered
	routes := router.Routes()
	foundSwagger := false
	foundAPIDocs := false

	for _, route := range routes {
		if route.Path == "/swagger/*any" {
			foundSwagger = true
		}
		if route.Path == "/api-docs" {
			foundAPIDocs = true
		}
	}

	assert.True(t, foundSwagger, "Swagger UI route not found")
	assert.True(t, foundAPIDocs, "API docs route not found")
}
