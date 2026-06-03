package repository

import (
	"errors"

	"github.com/basketikun/infinite-canvas/model"
	"gorm.io/gorm"
)

func SaveGeneratedImageRecord(item model.GeneratedImageRecord) (model.GeneratedImageRecord, error) {
	db, err := DB()
	if err != nil {
		return item, err
	}
	return item, db.Save(&item).Error
}

func GetGeneratedImageRecordByID(id string) (model.GeneratedImageRecord, bool, error) {
	db, err := DB()
	if err != nil {
		return model.GeneratedImageRecord{}, false, err
	}
	item := model.GeneratedImageRecord{}
	err = db.Where("id = ?", id).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.GeneratedImageRecord{}, false, nil
	}
	return item, err == nil, err
}

func ListGeneratedImageRecords(userID string, q model.Query) ([]model.GeneratedImageRecord, int64, error) {
	db, err := DB()
	if err != nil {
		return nil, 0, err
	}
	q.Normalize()
	tx := db.Model(&model.GeneratedImageRecord{}).Where("user_id = ?", userID)
	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		tx = tx.Where("prompt LIKE ? OR model LIKE ?", like, like)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.GeneratedImageRecord
	err = tx.Order("created_at desc").Offset(q.Offset()).Limit(q.PageSize).Find(&items).Error
	return items, total, err
}

func ListGalleryImages(q model.Query, admin bool) ([]model.GalleryImage, int64, error) {
	db, err := DB()
	if err != nil {
		return nil, 0, err
	}
	q.Normalize()
	tx := applyGalleryFilters(db.Model(&model.GalleryImage{}), q, admin)
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.GalleryImage
	err = tx.Order("recommended desc, created_at desc").Offset(q.Offset()).Limit(q.PageSize).Find(&items).Error
	return items, total, err
}

func ListGalleryTags(q model.Query, admin bool) ([]string, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	q.Normalize()
	q.Tags = nil
	var items []model.GalleryImage
	if err := applyGalleryFilters(db.Model(&model.GalleryImage{}), q, admin).Select("tags").Find(&items).Error; err != nil {
		return nil, err
	}
	return galleryTagsFromItems(items), nil
}

func SaveGalleryImage(item model.GalleryImage) (model.GalleryImage, error) {
	db, err := DB()
	if err != nil {
		return item, err
	}
	return item, db.Save(&item).Error
}

func GetGalleryImageByID(id string) (model.GalleryImage, bool, error) {
	db, err := DB()
	if err != nil {
		return model.GalleryImage{}, false, err
	}
	item := model.GalleryImage{}
	err = db.Where("id = ?", id).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.GalleryImage{}, false, nil
	}
	return item, err == nil, err
}

func GetGalleryImageByGeneratedID(id string) (model.GalleryImage, bool, error) {
	db, err := DB()
	if err != nil {
		return model.GalleryImage{}, false, err
	}
	item := model.GalleryImage{}
	err = db.Where("generated_image_id = ?", id).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.GalleryImage{}, false, nil
	}
	return item, err == nil, err
}

func applyGalleryFilters(tx *gorm.DB, q model.Query, admin bool) *gorm.DB {
	if !admin {
		tx = tx.Where("status = ?", model.GalleryStatusPublic)
	} else if isGalleryStatusOption(q.Type) {
		tx = tx.Where("status = ?", q.Type)
	}
	if q.Keyword != "" {
		like := "%" + q.Keyword + "%"
		if admin {
			tx = tx.Where("title LIKE ? OR description LIKE ? OR prompt LIKE ? OR model LIKE ?", like, like, like, like)
		} else {
			tx = tx.Where("title LIKE ? OR description LIKE ? OR model LIKE ? OR (show_prompt = ? AND prompt LIKE ?)", like, like, like, true, like)
		}
	}
	return applyGalleryTagsFilter(tx, q.Tags)
}

func applyGalleryTagsFilter(tx *gorm.DB, tags []string) *gorm.DB {
	if len(tags) == 0 {
		return tx
	}
	condition := tx.Session(&gorm.Session{NewDB: true})
	for _, tag := range tags {
		condition = condition.Or(assetJSONTagsContains(tx), tag)
	}
	return tx.Where(condition)
}

func galleryTagsFromItems(items []model.GalleryImage) []string {
	seen := map[string]bool{}
	tags := []string{}
	for _, item := range items {
		for _, tag := range item.Tags {
			if tag != "" && !seen[tag] {
				seen[tag] = true
				tags = append(tags, tag)
			}
		}
	}
	return tags
}

func isGalleryStatusOption(value string) bool {
	switch model.GalleryStatus(value) {
	case model.GalleryStatusPublic, model.GalleryStatusHidden, model.GalleryStatusDeleted:
		return true
	default:
		return false
	}
}
