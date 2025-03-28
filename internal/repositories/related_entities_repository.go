package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/pageza/alchemorsel-v1/internal/models"
	"gorm.io/gorm"
)

// CuisineRepository handles database operations for cuisines
type CuisineRepository interface {
	GetByID(ctx context.Context, id string) (*models.Cuisine, error)
	GetByName(ctx context.Context, name string) (*models.Cuisine, error)
	Create(ctx context.Context, cuisine *models.Cuisine) error
	List(ctx context.Context) ([]*models.Cuisine, error)
	Delete(ctx context.Context, id string) error
}

type DefaultCuisineRepository struct {
	db *gorm.DB
}

func NewCuisineRepository(db *gorm.DB) CuisineRepository {
	return &DefaultCuisineRepository{db: db}
}

func (r *DefaultCuisineRepository) GetByID(ctx context.Context, id string) (*models.Cuisine, error) {
	var cuisine models.Cuisine
	if err := r.db.WithContext(ctx).First(&cuisine, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &cuisine, nil
}

func (r *DefaultCuisineRepository) GetByName(ctx context.Context, name string) (*models.Cuisine, error) {
	var cuisine models.Cuisine
	if err := r.db.WithContext(ctx).First(&cuisine, "name = ?", name).Error; err != nil {
		return nil, err
	}
	return &cuisine, nil
}

func (r *DefaultCuisineRepository) Create(ctx context.Context, cuisine *models.Cuisine) error {
	if cuisine.ID == "" {
		cuisine.ID = uuid.New().String()
	}
	return r.db.WithContext(ctx).Create(cuisine).Error
}

func (r *DefaultCuisineRepository) List(ctx context.Context) ([]*models.Cuisine, error) {
	var cuisines []*models.Cuisine
	if err := r.db.WithContext(ctx).Find(&cuisines).Error; err != nil {
		return nil, err
	}
	return cuisines, nil
}

func (r *DefaultCuisineRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Cuisine{}, "id = ?", id).Error
}

// DietRepository handles database operations for diets
type DietRepository interface {
	GetByID(ctx context.Context, id string) (*models.Diet, error)
	GetByName(ctx context.Context, name string) (*models.Diet, error)
	Create(ctx context.Context, diet *models.Diet) error
	List(ctx context.Context) ([]*models.Diet, error)
	Delete(ctx context.Context, id string) error
}

type DefaultDietRepository struct {
	db *gorm.DB
}

func NewDietRepository(db *gorm.DB) DietRepository {
	return &DefaultDietRepository{db: db}
}

func (r *DefaultDietRepository) GetByID(ctx context.Context, id string) (*models.Diet, error) {
	var diet models.Diet
	if err := r.db.WithContext(ctx).First(&diet, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &diet, nil
}

func (r *DefaultDietRepository) GetByName(ctx context.Context, name string) (*models.Diet, error) {
	var diet models.Diet
	if err := r.db.WithContext(ctx).First(&diet, "name = ?", name).Error; err != nil {
		return nil, err
	}
	return &diet, nil
}

func (r *DefaultDietRepository) Create(ctx context.Context, diet *models.Diet) error {
	if diet.ID == "" {
		diet.ID = uuid.New().String()
	}
	return r.db.WithContext(ctx).Create(diet).Error
}

func (r *DefaultDietRepository) List(ctx context.Context) ([]*models.Diet, error) {
	var diets []*models.Diet
	if err := r.db.WithContext(ctx).Find(&diets).Error; err != nil {
		return nil, err
	}
	return diets, nil
}

func (r *DefaultDietRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Diet{}, "id = ?", id).Error
}

// ApplianceRepository handles database operations for appliances
type ApplianceRepository interface {
	GetByID(ctx context.Context, id string) (*models.Appliance, error)
	GetByName(ctx context.Context, name string) (*models.Appliance, error)
	Create(ctx context.Context, appliance *models.Appliance) error
	List(ctx context.Context) ([]*models.Appliance, error)
	Delete(ctx context.Context, id string) error
}

type DefaultApplianceRepository struct {
	db *gorm.DB
}

func NewApplianceRepository(db *gorm.DB) ApplianceRepository {
	return &DefaultApplianceRepository{db: db}
}

func (r *DefaultApplianceRepository) GetByID(ctx context.Context, id string) (*models.Appliance, error) {
	var appliance models.Appliance
	if err := r.db.WithContext(ctx).First(&appliance, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &appliance, nil
}

func (r *DefaultApplianceRepository) GetByName(ctx context.Context, name string) (*models.Appliance, error) {
	var appliance models.Appliance
	if err := r.db.WithContext(ctx).First(&appliance, "name = ?", name).Error; err != nil {
		return nil, err
	}
	return &appliance, nil
}

func (r *DefaultApplianceRepository) Create(ctx context.Context, appliance *models.Appliance) error {
	if appliance.ID == "" {
		appliance.ID = uuid.New().String()
	}
	return r.db.WithContext(ctx).Create(appliance).Error
}

func (r *DefaultApplianceRepository) List(ctx context.Context) ([]*models.Appliance, error) {
	var appliances []*models.Appliance
	if err := r.db.WithContext(ctx).Find(&appliances).Error; err != nil {
		return nil, err
	}
	return appliances, nil
}

func (r *DefaultApplianceRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Appliance{}, "id = ?", id).Error
}

// TagRepository handles database operations for tags
type TagRepository interface {
	GetByID(ctx context.Context, id string) (*models.Tag, error)
	GetByName(ctx context.Context, name string) (*models.Tag, error)
	Create(ctx context.Context, tag *models.Tag) error
	List(ctx context.Context) ([]*models.Tag, error)
	Delete(ctx context.Context, id string) error
}

type DefaultTagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) TagRepository {
	return &DefaultTagRepository{db: db}
}

func (r *DefaultTagRepository) GetByID(ctx context.Context, id string) (*models.Tag, error) {
	var tag models.Tag
	if err := r.db.WithContext(ctx).First(&tag, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *DefaultTagRepository) GetByName(ctx context.Context, name string) (*models.Tag, error) {
	var tag models.Tag
	if err := r.db.WithContext(ctx).First(&tag, "name = ?", name).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *DefaultTagRepository) Create(ctx context.Context, tag *models.Tag) error {
	if tag.ID == "" {
		tag.ID = uuid.New().String()
	}
	return r.db.WithContext(ctx).Create(tag).Error
}

func (r *DefaultTagRepository) List(ctx context.Context) ([]*models.Tag, error) {
	var tags []*models.Tag
	if err := r.db.WithContext(ctx).Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *DefaultTagRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Tag{}, "id = ?", id).Error
}
