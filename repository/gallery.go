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
	if err != nil {
		return items, total, err
	}
	items, err = fillGalleryCounts(db, items)
	return items, total, err
}

func MarkGalleryLiked(items []model.GalleryImage, userID string) ([]model.GalleryImage, error) {
	if len(items) == 0 || userID == "" {
		return items, nil
	}
	db, err := DB()
	if err != nil {
		return items, err
	}
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	var likes []model.GalleryLike
	if err := db.Where("user_id = ? AND gallery_id IN ?", userID, ids).Find(&likes).Error; err != nil {
		return items, err
	}
	liked := map[string]bool{}
	for _, item := range likes {
		liked[item.GalleryID] = true
	}
	for i := range items {
		items[i].Liked = liked[items[i].ID]
	}
	return items, nil
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

func DeleteGalleryImage(id string, generatedImageID string, now string) error {
	db, err := DB()
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&model.GalleryLike{}, "gallery_id = ?", id).Error; err != nil {
			return err
		}
		if err := tx.Delete(&model.GalleryComment{}, "gallery_id = ?", id).Error; err != nil {
			return err
		}
		if err := tx.Delete(&model.GalleryImage{}, "id = ?", id).Error; err != nil {
			return err
		}
		if generatedImageID == "" {
			return nil
		}
		return tx.Model(&model.GeneratedImageRecord{}).Where("id = ?", generatedImageID).Updates(map[string]any{
			"is_published": false,
			"updated_at":   now,
		}).Error
	})
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

func GetPublicGalleryImageByID(id string) (model.GalleryImage, bool, error) {
	db, err := DB()
	if err != nil {
		return model.GalleryImage{}, false, err
	}
	item := model.GalleryImage{}
	err = db.Where("id = ? AND status = ?", id, model.GalleryStatusPublic).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.GalleryImage{}, false, nil
	}
	return item, err == nil, err
}

func ToggleGalleryLike(galleryID string, userID string, now string) (model.GalleryImage, bool, error) {
	db, err := DB()
	if err != nil {
		return model.GalleryImage{}, false, err
	}
	liked := false
	var item model.GalleryImage
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND status = ?", galleryID, model.GalleryStatusPublic).First(&item).Error; err != nil {
			return err
		}
		var existing model.GalleryLike
		err := tx.Where("gallery_id = ? AND user_id = ?", galleryID, userID).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			liked = true
			if err := tx.Create(&model.GalleryLike{ID: userID + "-" + galleryID, GalleryID: galleryID, UserID: userID, CreatedAt: now}).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			liked = false
			if err := tx.Delete(&model.GalleryLike{}, "id = ?", existing.ID).Error; err != nil {
				return err
			}
		}
		if err := syncGalleryCounts(tx, galleryID, now); err != nil {
			return err
		}
		return tx.First(&item, "id = ?", galleryID).Error
	}); err != nil {
		return model.GalleryImage{}, false, err
	}
	item.Liked = liked
	return item, liked, nil
}

func ListGalleryComments(galleryID string, q model.Query) ([]model.GalleryComment, int64, error) {
	db, err := DB()
	if err != nil {
		return nil, 0, err
	}
	q.Normalize()
	tx := db.Model(&model.GalleryComment{}).Where("gallery_id = ? AND status = ?", galleryID, "public")
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.GalleryComment
	err = tx.Order("created_at desc").Offset(q.Offset()).Limit(q.PageSize).Find(&items).Error
	return items, total, err
}

func SaveGalleryComment(comment model.GalleryComment, now string) (model.GalleryComment, error) {
	db, err := DB()
	if err != nil {
		return comment, err
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		var image model.GalleryImage
		if err := tx.Where("id = ? AND status = ?", comment.GalleryID, model.GalleryStatusPublic).First(&image).Error; err != nil {
			return err
		}
		if err := tx.Create(&comment).Error; err != nil {
			return err
		}
		return syncGalleryCounts(tx, comment.GalleryID, now)
	}); err != nil {
		return comment, err
	}
	return comment, nil
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
	case model.GalleryStatusPublic, model.GalleryStatusHidden:
		return true
	default:
		return false
	}
}

type galleryCountRow struct {
	GalleryID string `gorm:"column:gallery_id"`
	Total     int64  `gorm:"column:total"`
}

func fillGalleryCounts(db *gorm.DB, items []model.GalleryImage) ([]model.GalleryImage, error) {
	if len(items) == 0 {
		return items, nil
	}
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	likeCounts, err := listGalleryLikeCounts(db, ids)
	if err != nil {
		return items, err
	}
	commentCounts, err := listGalleryCommentCounts(db, ids)
	if err != nil {
		return items, err
	}
	for i := range items {
		items[i].LikeCount = likeCounts[items[i].ID]
		items[i].CommentCount = commentCounts[items[i].ID]
	}
	return items, nil
}

func listGalleryLikeCounts(db *gorm.DB, ids []string) (map[string]int, error) {
	var rows []galleryCountRow
	if err := db.Model(&model.GalleryLike{}).Select("gallery_id, COUNT(*) AS total").Where("gallery_id IN ?", ids).Group("gallery_id").Scan(&rows).Error; err != nil {
		return nil, err
	}
	return galleryCountMap(rows), nil
}

func listGalleryCommentCounts(db *gorm.DB, ids []string) (map[string]int, error) {
	var rows []galleryCountRow
	if err := db.Model(&model.GalleryComment{}).Select("gallery_id, COUNT(*) AS total").Where("gallery_id IN ? AND status = ?", ids, "public").Group("gallery_id").Scan(&rows).Error; err != nil {
		return nil, err
	}
	return galleryCountMap(rows), nil
}

func galleryCountMap(rows []galleryCountRow) map[string]int {
	counts := map[string]int{}
	for _, row := range rows {
		counts[row.GalleryID] = int(row.Total)
	}
	return counts
}

func syncGalleryCounts(tx *gorm.DB, galleryID string, now string) error {
	likeCount, err := countGalleryLikes(tx, galleryID)
	if err != nil {
		return err
	}
	commentCount, err := countGalleryComments(tx, galleryID)
	if err != nil {
		return err
	}
	return tx.Model(&model.GalleryImage{}).Where("id = ?", galleryID).Updates(map[string]any{
		"like_count":    likeCount,
		"comment_count": commentCount,
		"updated_at":    now,
	}).Error
}

func countGalleryLikes(tx *gorm.DB, galleryID string) (int, error) {
	var count int64
	err := tx.Model(&model.GalleryLike{}).Where("gallery_id = ?", galleryID).Count(&count).Error
	return int(count), err
}

func countGalleryComments(tx *gorm.DB, galleryID string) (int, error) {
	var count int64
	err := tx.Model(&model.GalleryComment{}).Where("gallery_id = ? AND status = ?", galleryID, "public").Count(&count).Error
	return int(count), err
}
