package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pageza/alchemorsel-v1/internal/dtos"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/services"
)

// RelatedEntityRequest represents the request body for creating/updating related entities
type RelatedEntityRequest struct {
	Name string `json:"name" binding:"required"`
}

// RelatedEntityResponse represents the response body for related entities
type RelatedEntityResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CuisineHandler handles HTTP requests for cuisines
type CuisineHandler struct {
	service services.CuisineService
}

func NewCuisineHandler(service services.CuisineService) *CuisineHandler {
	return &CuisineHandler{service: service}
}

func (h *CuisineHandler) RegisterRoutes(r *gin.RouterGroup) {
	cuisines := r.Group("/cuisines")
	{
		cuisines.GET("", h.List)
		cuisines.POST("", h.Create)
		cuisines.GET("/:id", h.GetByID)
		cuisines.DELETE("/:id", h.Delete)
	}
}

func (h *CuisineHandler) List(c *gin.Context) {
	cuisines, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to list cuisines",
		})
		return
	}

	response := make([]RelatedEntityResponse, len(cuisines))
	for i, cuisine := range cuisines {
		response[i] = RelatedEntityResponse{
			ID:   cuisine.ID,
			Name: cuisine.Name,
		}
	}

	c.JSON(http.StatusOK, response)
}

func (h *CuisineHandler) Create(c *gin.Context) {
	var req RelatedEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
		return
	}

	cuisine := &models.Cuisine{Name: req.Name}
	if err := h.service.Create(c.Request.Context(), cuisine); err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to create cuisine",
		})
		return
	}

	response := RelatedEntityResponse{
		ID:   cuisine.ID,
		Name: cuisine.Name,
	}

	c.JSON(http.StatusCreated, response)
}

func (h *CuisineHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	cuisine, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "Cuisine not found",
		})
		return
	}

	response := RelatedEntityResponse{
		ID:   cuisine.ID,
		Name: cuisine.Name,
	}

	c.JSON(http.StatusOK, response)
}

func (h *CuisineHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to delete cuisine",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// DietHandler handles HTTP requests for diets
type DietHandler struct {
	service services.DietService
}

func NewDietHandler(service services.DietService) *DietHandler {
	return &DietHandler{service: service}
}

func (h *DietHandler) RegisterRoutes(r *gin.RouterGroup) {
	diets := r.Group("/diets")
	{
		diets.GET("", h.List)
		diets.POST("", h.Create)
		diets.GET("/:id", h.GetByID)
		diets.DELETE("/:id", h.Delete)
	}
}

func (h *DietHandler) List(c *gin.Context) {
	diets, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to list diets",
		})
		return
	}

	response := make([]RelatedEntityResponse, len(diets))
	for i, diet := range diets {
		response[i] = RelatedEntityResponse{
			ID:   diet.ID,
			Name: diet.Name,
		}
	}

	c.JSON(http.StatusOK, response)
}

func (h *DietHandler) Create(c *gin.Context) {
	var req RelatedEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
		return
	}

	diet := &models.Diet{Name: req.Name}
	if err := h.service.Create(c.Request.Context(), diet); err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to create diet",
		})
		return
	}

	response := RelatedEntityResponse{
		ID:   diet.ID,
		Name: diet.Name,
	}

	c.JSON(http.StatusCreated, response)
}

func (h *DietHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	diet, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "Diet not found",
		})
		return
	}

	response := RelatedEntityResponse{
		ID:   diet.ID,
		Name: diet.Name,
	}

	c.JSON(http.StatusOK, response)
}

func (h *DietHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to delete diet",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// ApplianceHandler handles HTTP requests for appliances
type ApplianceHandler struct {
	service services.ApplianceService
}

func NewApplianceHandler(service services.ApplianceService) *ApplianceHandler {
	return &ApplianceHandler{service: service}
}

func (h *ApplianceHandler) RegisterRoutes(r *gin.RouterGroup) {
	appliances := r.Group("/appliances")
	{
		appliances.GET("", h.List)
		appliances.POST("", h.Create)
		appliances.GET("/:id", h.GetByID)
		appliances.DELETE("/:id", h.Delete)
	}
}

func (h *ApplianceHandler) List(c *gin.Context) {
	appliances, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to list appliances",
		})
		return
	}

	response := make([]RelatedEntityResponse, len(appliances))
	for i, appliance := range appliances {
		response[i] = RelatedEntityResponse{
			ID:   appliance.ID,
			Name: appliance.Name,
		}
	}

	c.JSON(http.StatusOK, response)
}

func (h *ApplianceHandler) Create(c *gin.Context) {
	var req RelatedEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
		return
	}

	appliance := &models.Appliance{Name: req.Name}
	if err := h.service.Create(c.Request.Context(), appliance); err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to create appliance",
		})
		return
	}

	response := RelatedEntityResponse{
		ID:   appliance.ID,
		Name: appliance.Name,
	}

	c.JSON(http.StatusCreated, response)
}

func (h *ApplianceHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	appliance, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "Appliance not found",
		})
		return
	}

	response := RelatedEntityResponse{
		ID:   appliance.ID,
		Name: appliance.Name,
	}

	c.JSON(http.StatusOK, response)
}

func (h *ApplianceHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to delete appliance",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// TagHandler handles HTTP requests for tags
type TagHandler struct {
	service services.TagService
}

func NewTagHandler(service services.TagService) *TagHandler {
	return &TagHandler{service: service}
}

func (h *TagHandler) RegisterRoutes(r *gin.RouterGroup) {
	tags := r.Group("/tags")
	{
		tags.GET("", h.List)
		tags.POST("", h.Create)
		tags.GET("/:id", h.GetByID)
		tags.DELETE("/:id", h.Delete)
	}
}

func (h *TagHandler) List(c *gin.Context) {
	tags, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to list tags",
		})
		return
	}

	response := make([]RelatedEntityResponse, len(tags))
	for i, tag := range tags {
		response[i] = RelatedEntityResponse{
			ID:   tag.ID,
			Name: tag.Name,
		}
	}

	c.JSON(http.StatusOK, response)
}

func (h *TagHandler) Create(c *gin.Context) {
	var req RelatedEntityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
		return
	}

	tag := &models.Tag{Name: req.Name}
	if err := h.service.Create(c.Request.Context(), tag); err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to create tag",
		})
		return
	}

	response := RelatedEntityResponse{
		ID:   tag.ID,
		Name: tag.Name,
	}

	c.JSON(http.StatusCreated, response)
}

func (h *TagHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	tag, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "Tag not found",
		})
		return
	}

	response := RelatedEntityResponse{
		ID:   tag.ID,
		Name: tag.Name,
	}

	c.JSON(http.StatusOK, response)
}

func (h *TagHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to delete tag",
		})
		return
	}

	c.Status(http.StatusNoContent)
}
