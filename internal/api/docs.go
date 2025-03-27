package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// APIDoc represents the OpenAPI documentation
type APIDoc struct {
	Info       Info       `json:"info"`
	Servers    []Server   `json:"servers"`
	Paths      Paths      `json:"paths"`
	Components Components `json:"components"`
}

// Info contains API metadata
type Info struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Contact     struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"contact"`
	License struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"license"`
}

// Server represents an API server
type Server struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

// Paths contains all API endpoints
type Paths map[string]PathItem

// PathItem represents an API endpoint
type PathItem struct {
	Get     *Operation `json:"get,omitempty"`
	Post    *Operation `json:"post,omitempty"`
	Put     *Operation `json:"put,omitempty"`
	Delete  *Operation `json:"delete,omitempty"`
	Patch   *Operation `json:"patch,omitempty"`
	Options *Operation `json:"options,omitempty"`
	Head    *Operation `json:"head,omitempty"`
}

// Operation represents an HTTP operation
type Operation struct {
	Summary     string                `json:"summary"`
	Description string                `json:"description"`
	Tags        []string              `json:"tags"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Security    []map[string][]string `json:"security,omitempty"`
}

// Parameter represents an API parameter
type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Schema      Schema `json:"schema"`
}

// RequestBody represents a request body
type RequestBody struct {
	Description string               `json:"description"`
	Required    bool                 `json:"required"`
	Content     map[string]MediaType `json:"content"`
}

// Response represents an API response
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// MediaType represents a media type
type MediaType struct {
	Schema Schema `json:"schema"`
}

// Schema represents a JSON Schema
type Schema struct {
	Type       string            `json:"type,omitempty"`
	Format     string            `json:"format,omitempty"`
	Properties map[string]Schema `json:"properties,omitempty"`
	Required   []string          `json:"required,omitempty"`
	Items      *Schema           `json:"items,omitempty"`
	Ref        string            `json:"$ref,omitempty"`
}

// Components contains reusable components
type Components struct {
	Schemas         map[string]Schema         `json:"schemas,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

// SecurityScheme represents a security scheme
type SecurityScheme struct {
	Type         string      `json:"type"`
	Description  string      `json:"description,omitempty"`
	Scheme       string      `json:"scheme,omitempty"`
	BearerFormat string      `json:"bearerFormat,omitempty"`
	Flows        *OAuthFlows `json:"flows,omitempty"`
}

// OAuthFlows represents OAuth flows
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty"`
}

// OAuthFlow represents an OAuth flow
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty"`
	Scopes           map[string]string `json:"scopes,omitempty"`
}

// APIDocManager handles API documentation
type APIDocManager struct {
	doc     *APIDoc
	version string
}

// NewAPIDocManager creates a new API documentation manager
func NewAPIDocManager(version string) *APIDocManager {
	return &APIDocManager{
		version: version,
		doc: &APIDoc{
			Info: Info{
				Title:       "Alchemorsel API",
				Description: "API for the Alchemorsel application",
				Version:     version,
				Contact: struct {
					Name  string `json:"name"`
					Email string `json:"email"`
				}{
					Name:  "Alchemorsel Team",
					Email: "support@alchemorsel.com",
				},
				License: struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				}{
					Name: "MIT",
					URL:  "https://opensource.org/licenses/MIT",
				},
			},
			Servers: []Server{
				{
					URL:         "/api/v" + version,
					Description: "API v" + version,
				},
			},
			Paths: make(Paths),
			Components: Components{
				Schemas:         make(map[string]Schema),
				SecuritySchemes: make(map[string]SecurityScheme),
			},
		},
	}
}

// AddPath adds a new API path
func (m *APIDocManager) AddPath(path string, item PathItem) {
	m.doc.Paths[path] = item
}

// AddSchema adds a new schema component
func (m *APIDocManager) AddSchema(name string, schema Schema) {
	m.doc.Components.Schemas[name] = schema
}

// AddSecurityScheme adds a new security scheme
func (m *APIDocManager) AddSecurityScheme(name string, scheme SecurityScheme) {
	m.doc.Components.SecuritySchemes[name] = scheme
}

// GetDoc returns the API documentation
func (m *APIDocManager) GetDoc() *APIDoc {
	return m.doc
}

// RegisterSwagger registers Swagger UI routes
func (m *APIDocManager) RegisterSwagger(r *gin.Engine) {
	// Serve Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Serve OpenAPI spec
	r.GET("/api-docs", func(c *gin.Context) {
		c.JSON(http.StatusOK, m.doc)
	})
}

// GenerateMarkdown generates markdown documentation
func (m *APIDocManager) GenerateMarkdown() (string, error) {
	var sb strings.Builder

	// Write header
	sb.WriteString(fmt.Sprintf("# %s\n\n", m.doc.Info.Title))
	sb.WriteString(fmt.Sprintf("%s\n\n", m.doc.Info.Description))
	sb.WriteString(fmt.Sprintf("Version: %s\n\n", m.doc.Info.Version))

	// Write paths
	sb.WriteString("## Endpoints\n\n")
	for path, item := range m.doc.Paths {
		sb.WriteString(fmt.Sprintf("### %s\n\n", path))

		if item.Get != nil {
			sb.WriteString("#### GET\n")
			sb.WriteString(fmt.Sprintf("%s\n\n", item.Get.Description))
		}

		if item.Post != nil {
			sb.WriteString("#### POST\n")
			sb.WriteString(fmt.Sprintf("%s\n\n", item.Post.Description))
		}

		if item.Put != nil {
			sb.WriteString("#### PUT\n")
			sb.WriteString(fmt.Sprintf("%s\n\n", item.Put.Description))
		}

		if item.Delete != nil {
			sb.WriteString("#### DELETE\n")
			sb.WriteString(fmt.Sprintf("%s\n\n", item.Delete.Description))
		}
	}

	return sb.String(), nil
}
