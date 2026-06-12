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
	err = tx.Order(galleryImageOrder(q.Sort)).Offset(q.Offset()).Limit(q.PageSize).Find(&items).Error
	if err != nil {
		return items, total, err
	}
	items, err = fillGalleryCounts(db, items)
	if err != nil {
		return items, total, err
	}
	items, err = fillGalleryImageUsers(db, items)
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

func CreateGalleryImageWithFile(item model.GalleryImage, file model.GalleryImageFile, record model.GeneratedImageRecord) (model.GalleryImage, error) {
	db, err := DB()
	if err != nil {
		return item, err
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&item).Error; err != nil {
			return err
		}
		if err := tx.Create(&file).Error; err != nil {
			return err
		}
		return tx.Save(&record).Error
	})
	if err == nil {
		cacheGalleryImageFile(file)
	}
	return item, err
}

func GetGalleryImageFileByGalleryID(galleryID string) (model.GalleryImageFile, bool, error) {
	if item, ok := cachedGalleryImageFile(galleryID); ok {
		return item, true, nil
	}
	db, err := DB()
	if err != nil {
		return model.GalleryImageFile{}, false, err
	}
	item := model.GalleryImageFile{}
	err = db.Where("gallery_id = ?", galleryID).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.GalleryImageFile{}, false, nil
	}
	if err == nil {
		cacheGalleryImageFile(item)
	}
	return item, err == nil, err
}

func DeleteGalleryImage(id string, generatedImageID string, now string) error {
	db, err := DB()
	if err != nil {
		return err
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&model.GalleryImageFile{}, "gallery_id = ?", id).Error; err != nil {
			return err
		}
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
	if err == nil {
		CacheDelete(galleryImageFileCacheKey(id))
	}
	return err
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
	items, err := fillGalleryImageUsers(db, []model.GalleryImage{item})
	if err != nil {
		return item, liked, err
	}
	return items[0], liked, nil
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
	if err != nil {
		return items, total, err
	}
	items, err = fillGalleryCommentUsers(db, items)
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

func galleryImageOrder(sort string) string {
	switch sort {
	case "likes":
		return "recommended desc, like_count desc, created_at desc"
	default:
		return "recommended desc, created_at desc"
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

func fillGalleryImageUsers(db *gorm.DB, items []model.GalleryImage) ([]model.GalleryImage, error) {
	if len(items) == 0 {
		return items, nil
	}
	users, err := galleryUsersByID(db, galleryImageUserIDs(items))
	if err != nil {
		return items, err
	}
	for i := range items {
		user := users[items[i].UserID]
		items[i].Username = user.Username
		items[i].DisplayName = user.DisplayName
		items[i].AvatarURL = user.AvatarURL
	}
	return items, nil
}

func fillGalleryCommentUsers(db *gorm.DB, items []model.GalleryComment) ([]model.GalleryComment, error) {
	if len(items) == 0 {
		return items, nil
	}
	users, err := galleryUsersByID(db, galleryCommentUserIDs(items))
	if err != nil {
		return items, err
	}
	for i := range items {
		if user, ok := users[items[i].UserID]; ok {
			items[i].Username = firstNonEmptyString(user.Username, items[i].Username)
			items[i].DisplayName = firstNonEmptyString(user.DisplayName, items[i].DisplayName)
			items[i].AvatarURL = firstNonEmptyString(user.AvatarURL, items[i].AvatarURL)
		}
	}
	return items, nil
}

func galleryUsersByID(db *gorm.DB, ids []string) (map[string]model.User, error) {
	if len(ids) == 0 {
		return map[string]model.User{}, nil
	}
	var users []model.User
	if err := db.Select("id", "username", "display_name", "avatar_url").Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	byID := make(map[string]model.User, len(users))
	for _, user := range users {
		byID[user.ID] = user
	}
	return byID, nil
}

func galleryImageUserIDs(items []model.GalleryImage) []string {
	seen := map[string]bool{}
	ids := []string{}
	for _, item := range items {
		if item.UserID != "" && !seen[item.UserID] {
			seen[item.UserID] = true
			ids = append(ids, item.UserID)
		}
	}
	return ids
}

func galleryCommentUserIDs(items []model.GalleryComment) []string {
	seen := map[string]bool{}
	ids := []string{}
	for _, item := range items {
		if item.UserID != "" && !seen[item.UserID] {
			seen[item.UserID] = true
			ids = append(ids, item.UserID)
		}
	}
	return ids
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
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

func galleryImageFileCacheKey(galleryID string) string {
	return CacheKey("gallery-image-file", galleryID)
}

func cachedGalleryImageFile(galleryID string) (model.GalleryImageFile, bool) {
	client := Redis()
	if client == nil || galleryID == "" {
		return model.GalleryImageFile{}, false
	}
	ctx := cacheContext()
	key := galleryImageFileCacheKey(galleryID)
	pipe := client.Pipeline()
	id := pipe.HGet(ctx, key, "id")
	sourceURL := pipe.HGet(ctx, key, "source_url")
	mimeType := pipe.HGet(ctx, key, "mime_type")
	createdAt := pipe.HGet(ctx, key, "created_at")
	updatedAt := pipe.HGet(ctx, key, "updated_at")
	data := pipe.HGet(ctx, key, "data")
	if _, err := pipe.Exec(ctx); err != nil {
		return model.GalleryImageFile{}, false
	}
	bytes, err := data.Bytes()
	if err != nil || len(bytes) == 0 {
		return model.GalleryImageFile{}, false
	}
	return model.GalleryImageFile{
		ID:        id.Val(),
		GalleryID: galleryID,
		SourceURL: sourceURL.Val(),
		MimeType:  mimeType.Val(),
		Size:      int64(len(bytes)),
		Data:      bytes,
		CreatedAt: createdAt.Val(),
		UpdatedAt: updatedAt.Val(),
	}, true
}

func cacheGalleryImageFile(file model.GalleryImageFile) {
	client := Redis()
	if client == nil || file.GalleryID == "" || len(file.Data) == 0 {
		return
	}
	ctx := cacheContext()
	key := galleryImageFileCacheKey(file.GalleryID)
	pipe := client.Pipeline()
	pipe.HSet(ctx, key, map[string]any{
		"id":         file.ID,
		"source_url": file.SourceURL,
		"mime_type":  file.MimeType,
		"created_at": file.CreatedAt,
		"updated_at": file.UpdatedAt,
		"data":       file.Data,
	})
	pipe.Expire(ctx, key, CacheTTL())
	_, _ = pipe.Exec(ctx)
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
