package postgres

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/Wei-Shaw/sub2api/backend/internal/model"
	"github.com/Wei-Shaw/sub2api/backend/internal/repository"
)

type adminRepository struct {
	db *gorm.DB
}

// NewAdminRepository creates a new admin repository.
func NewAdminRepository(db *gorm.DB) repository.AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) Create(ctx context.Context, admin *model.Admin) error {
	if err := r.db.WithContext(ctx).Create(admin).Error; err != nil {
		return fmt.Errorf("failed to create admin: %w", err)
	}
	return nil
}

func (r *adminRepository) GetByID(ctx context.Context, id int64) (*model.Admin, error) {
	var admin model.Admin
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&admin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("admin not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get admin by id: %w", err)
	}
	return &admin, nil
}

func (r *adminRepository) GetByUsername(ctx context.Context, username string) (*model.Admin, error) {
	var admin model.Admin
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&admin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("admin not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get admin by username: %w", err)
	}
	return &admin, nil
}

func (r *adminRepository) List(ctx context.Context, offset, limit int) ([]*model.Admin, error) {
	var admins []*model.Admin
	query := r.db.WithContext(ctx).Order("id DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&admins).Error; err != nil {
		return nil, fmt.Errorf("failed to list admins: %w", err)
	}
	return admins, nil
}

func (r *adminRepository) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&model.Admin{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update admin: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("admin not found: id=%d", id)
	}

	return nil
}

func (r *adminRepository) Delete(ctx context.Context, id int64) error {
	if err := r.db.WithContext(ctx).Delete(&model.Admin{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete admin: %w", err)
	}
	return nil
}
