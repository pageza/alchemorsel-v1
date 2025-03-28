package services

import (
	"context"
	"fmt"

	"github.com/pageza/alchemorsel-v1/internal/models"
	"github.com/pageza/alchemorsel-v1/internal/repositories"
)

// CuisineService handles business logic for cuisines
type CuisineService interface {
	GetByID(ctx context.Context, id string) (*models.Cuisine, error)
	GetByName(ctx context.Context, name string) (*models.Cuisine, error)
	Create(ctx context.Context, cuisine *models.Cuisine) error
	List(ctx context.Context) ([]*models.Cuisine, error)
	Delete(ctx context.Context, id string) error
	GetOrCreate(ctx context.Context, name string) (*models.Cuisine, error)
}

type DefaultCuisineService struct {
	repo repositories.CuisineRepository
}

func NewCuisineService(repo repositories.CuisineRepository) CuisineService {
	return &DefaultCuisineService{repo: repo}
}

func (s *DefaultCuisineService) GetByID(ctx context.Context, id string) (*models.Cuisine, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DefaultCuisineService) GetByName(ctx context.Context, name string) (*models.Cuisine, error) {
	return s.repo.GetByName(ctx, name)
}

func (s *DefaultCuisineService) Create(ctx context.Context, cuisine *models.Cuisine) error {
	if cuisine.Name == "" {
		return fmt.Errorf("cuisine name is required")
	}
	return s.repo.Create(ctx, cuisine)
}

func (s *DefaultCuisineService) List(ctx context.Context) ([]*models.Cuisine, error) {
	return s.repo.List(ctx)
}

func (s *DefaultCuisineService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *DefaultCuisineService) GetOrCreate(ctx context.Context, name string) (*models.Cuisine, error) {
	cuisine, err := s.repo.GetByName(ctx, name)
	if err == nil {
		return cuisine, nil
	}

	cuisine = &models.Cuisine{Name: name}
	err = s.repo.Create(ctx, cuisine)
	if err != nil {
		return nil, err
	}
	return cuisine, nil
}

// DietService handles business logic for diets
type DietService interface {
	GetByID(ctx context.Context, id string) (*models.Diet, error)
	GetByName(ctx context.Context, name string) (*models.Diet, error)
	Create(ctx context.Context, diet *models.Diet) error
	List(ctx context.Context) ([]*models.Diet, error)
	Delete(ctx context.Context, id string) error
	GetOrCreate(ctx context.Context, name string) (*models.Diet, error)
}

type DefaultDietService struct {
	repo repositories.DietRepository
}

func NewDietService(repo repositories.DietRepository) DietService {
	return &DefaultDietService{repo: repo}
}

func (s *DefaultDietService) GetByID(ctx context.Context, id string) (*models.Diet, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DefaultDietService) GetByName(ctx context.Context, name string) (*models.Diet, error) {
	return s.repo.GetByName(ctx, name)
}

func (s *DefaultDietService) Create(ctx context.Context, diet *models.Diet) error {
	if diet.Name == "" {
		return fmt.Errorf("diet name is required")
	}
	return s.repo.Create(ctx, diet)
}

func (s *DefaultDietService) List(ctx context.Context) ([]*models.Diet, error) {
	return s.repo.List(ctx)
}

func (s *DefaultDietService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *DefaultDietService) GetOrCreate(ctx context.Context, name string) (*models.Diet, error) {
	diet, err := s.repo.GetByName(ctx, name)
	if err == nil {
		return diet, nil
	}

	diet = &models.Diet{Name: name}
	err = s.repo.Create(ctx, diet)
	if err != nil {
		return nil, err
	}
	return diet, nil
}

// ApplianceService handles business logic for appliances
type ApplianceService interface {
	GetByID(ctx context.Context, id string) (*models.Appliance, error)
	GetByName(ctx context.Context, name string) (*models.Appliance, error)
	Create(ctx context.Context, appliance *models.Appliance) error
	List(ctx context.Context) ([]*models.Appliance, error)
	Delete(ctx context.Context, id string) error
	GetOrCreate(ctx context.Context, name string) (*models.Appliance, error)
}

type DefaultApplianceService struct {
	repo repositories.ApplianceRepository
}

func NewApplianceService(repo repositories.ApplianceRepository) ApplianceService {
	return &DefaultApplianceService{repo: repo}
}

func (s *DefaultApplianceService) GetByID(ctx context.Context, id string) (*models.Appliance, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DefaultApplianceService) GetByName(ctx context.Context, name string) (*models.Appliance, error) {
	return s.repo.GetByName(ctx, name)
}

func (s *DefaultApplianceService) Create(ctx context.Context, appliance *models.Appliance) error {
	if appliance.Name == "" {
		return fmt.Errorf("appliance name is required")
	}
	return s.repo.Create(ctx, appliance)
}

func (s *DefaultApplianceService) List(ctx context.Context) ([]*models.Appliance, error) {
	return s.repo.List(ctx)
}

func (s *DefaultApplianceService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *DefaultApplianceService) GetOrCreate(ctx context.Context, name string) (*models.Appliance, error) {
	appliance, err := s.repo.GetByName(ctx, name)
	if err == nil {
		return appliance, nil
	}

	appliance = &models.Appliance{Name: name}
	err = s.repo.Create(ctx, appliance)
	if err != nil {
		return nil, err
	}
	return appliance, nil
}

// TagService handles business logic for tags
type TagService interface {
	GetByID(ctx context.Context, id string) (*models.Tag, error)
	GetByName(ctx context.Context, name string) (*models.Tag, error)
	Create(ctx context.Context, tag *models.Tag) error
	List(ctx context.Context) ([]*models.Tag, error)
	Delete(ctx context.Context, id string) error
	GetOrCreate(ctx context.Context, name string) (*models.Tag, error)
}

type DefaultTagService struct {
	repo repositories.TagRepository
}

func NewTagService(repo repositories.TagRepository) TagService {
	return &DefaultTagService{repo: repo}
}

func (s *DefaultTagService) GetByID(ctx context.Context, id string) (*models.Tag, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DefaultTagService) GetByName(ctx context.Context, name string) (*models.Tag, error) {
	return s.repo.GetByName(ctx, name)
}

func (s *DefaultTagService) Create(ctx context.Context, tag *models.Tag) error {
	if tag.Name == "" {
		return fmt.Errorf("tag name is required")
	}
	return s.repo.Create(ctx, tag)
}

func (s *DefaultTagService) List(ctx context.Context) ([]*models.Tag, error) {
	return s.repo.List(ctx)
}

func (s *DefaultTagService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *DefaultTagService) GetOrCreate(ctx context.Context, name string) (*models.Tag, error) {
	tag, err := s.repo.GetByName(ctx, name)
	if err == nil {
		return tag, nil
	}

	tag = &models.Tag{Name: name}
	err = s.repo.Create(ctx, tag)
	if err != nil {
		return nil, err
	}
	return tag, nil
}
