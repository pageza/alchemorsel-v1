package api_test

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/api"
	"github.com/stretchr/testify/assert"
)

func TestNewAPIDocManager(t *testing.T) {
	version := "1.0.0"
	manager := api.NewAPIDocManager(version)

	assert.NotNil(t, manager)
	assert.Equal(t, version, manager.GetVersion())
	assert.NotNil(t, manager.GetDoc())

	// Check basic info
	assert.Equal(t, "Alchemorsel API", manager.GetDoc().Info.Title)
	assert.Equal(t, "API for the Alchemorsel application", manager.GetDoc().Info.Description)
	assert.Equal(t, version, manager.GetDoc().Info.Version)
	assert.Equal(t, "Alchemorsel Team", manager.GetDoc().Info.Contact.Name)
	assert.Equal(t, "support@alchemorsel.com", manager.GetDoc().Info.Contact.Email)
	assert.Equal(t, "MIT", manager.GetDoc().Info.License.Name)
	assert.Equal(t, "https://opensource.org/licenses/MIT", manager.GetDoc().Info.License.URL)

	// Check servers
	assert.Len(t, manager.GetDoc().Servers, 1)
	assert.Equal(t, "/api/v"+version, manager.GetDoc().Servers[0].URL)
	assert.Equal(t, "API v"+version, manager.GetDoc().Servers[0].Description)

	// Check components
	assert.NotNil(t, manager.GetDoc().Components.Schemas)
	assert.NotNil(t, manager.GetDoc().Components.SecuritySchemes)
}

func TestAddPath(t *testing.T) {
	manager := api.NewAPIDocManager("1.0.0")

	// Create a test path item
	pathItem := api.PathItem{
		Get: &api.Operation{
			Summary:     "Get user",
			Description: "Retrieve user information",
			Tags:        []string{"users"},
			Responses: map[string]api.Response{
				"200": {
					Description: "Success",
					Content: map[string]api.MediaType{
						"application/json": {
							Schema: api.Schema{
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
	assert.Contains(t, manager.GetDoc().Paths, path)
	assert.Equal(t, pathItem, manager.GetDoc().Paths[path])
}

func TestAddSchema(t *testing.T) {
	manager := api.NewAPIDocManager("1.0.0")

	// Create a test schema
	schema := api.Schema{
		Type: "object",
		Properties: map[string]api.Schema{
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
	assert.Contains(t, manager.GetDoc().Components.Schemas, schemaName)
	assert.Equal(t, schema, manager.GetDoc().Components.Schemas[schemaName])
}

func TestAddSecurityScheme(t *testing.T) {
	manager := api.NewAPIDocManager("1.0.0")

	// Create a test security scheme
	scheme := api.SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Description:  "JWT token authentication",
	}

	// Add the security scheme
	schemeName := "bearerAuth"
	manager.AddSecurityScheme(schemeName, scheme)

	// Verify the security scheme was added
	assert.Contains(t, manager.GetDoc().Components.SecuritySchemes, schemeName)
	assert.Equal(t, scheme, manager.GetDoc().Components.SecuritySchemes[schemeName])
}

func TestGenerateMarkdown(t *testing.T) {
	manager := api.NewAPIDocManager("1.0.0")

	// Add a test path
	pathItem := api.PathItem{
		Get: &api.Operation{
			Summary:     "Get user",
			Description: "Retrieve user information",
			Tags:        []string{"users"},
		},
		Post: &api.Operation{
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
	manager := api.NewAPIDocManager("1.0.0")
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
