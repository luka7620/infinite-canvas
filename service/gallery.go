package service

import (
	"strings"

	"github.com/basketikun/infinite-canvas/model"
	"github.com/basketikun/infinite-canvas/repository"
)

type CreateGeneratedImageRecordInput struct {
	UserID   string `json:"userId"`
	ImageURL string `json:"imageUrl"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	MimeType string `json:"mimeType"`
	Model    string `json:"model"`
	Prompt   string `json:"prompt"`
	Source   string `json:"source"`
}

type PublishGalleryImageInput struct {
	GeneratedImageID string   `json:"generatedImageId"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Tags             []string `json:"tags"`
	ShowPrompt       bool     `json:"showPrompt"`
}

type UpdateGalleryImageInput struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Tags        []string            `json:"tags"`
	ShowPrompt  bool                `json:"showPrompt"`
	Status      model.GalleryStatus `json:"status"`
	Recommended bool                `json:"recommended"`
}

func CreateGeneratedImageRecord(input CreateGeneratedImageRecordInput) (model.GeneratedImageRecord, error) {
	now := now()
	record := model.GeneratedImageRecord{
		ID:        newID("genimg"),
		UserID:    strings.TrimSpace(input.UserID),
		ImageURL:  strings.TrimSpace(input.ImageURL),
		Width:     input.Width,
		Height:    input.Height,
		MimeType:  strings.TrimSpace(input.MimeType),
		Model:     strings.TrimSpace(input.Model),
		Prompt:    strings.TrimSpace(input.Prompt),
		Source:    normalizeGeneratedImageSource(input.Source),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if record.UserID == "" {
		return record, safeMessageError{message: "未登录或权限不足"}
	}
	if record.ImageURL == "" {
		return record, safeMessageError{message: "图片地址不能为空"}
	}
	return repository.SaveGeneratedImageRecord(record)
}

func ListMyGeneratedImageRecords(userID string, q model.Query) (model.GeneratedImageRecordList, error) {
	items, total, err := repository.ListGeneratedImageRecords(userID, q)
	if err != nil {
		return model.GeneratedImageRecordList{}, err
	}
	return model.GeneratedImageRecordList{Items: items, Total: int(total)}, nil
}

func PublishGalleryImage(user model.AuthUser, input PublishGalleryImageInput) (model.GalleryImage, error) {
	record, ok, err := repository.GetGeneratedImageRecordByID(strings.TrimSpace(input.GeneratedImageID))
	if err != nil || !ok {
		if err != nil {
			return model.GalleryImage{}, err
		}
		return model.GalleryImage{}, safeMessageError{message: "生成图片记录不存在"}
	}
	if record.UserID != user.ID {
		return model.GalleryImage{}, safeMessageError{message: "只能发布自己的生成图片"}
	}
	if !isAllowedGeneratedImageSource(record.Source) {
		return model.GalleryImage{}, safeMessageError{message: "该图片来源不支持发布到画廊"}
	}
	if record.IsPublished {
		return model.GalleryImage{}, safeMessageError{message: "该图片已发布"}
	}
	if _, ok, err := repository.GetGalleryImageByGeneratedID(record.ID); err != nil || ok {
		if err != nil {
			return model.GalleryImage{}, err
		}
		return model.GalleryImage{}, safeMessageError{message: "该图片已发布"}
	}
	now := now()
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = firstNonEmpty(strings.TrimSpace(record.Prompt), "未命名作品")
	}
	title = truncateRunes(title, 60)
	item := model.GalleryImage{
		ID:               newID("gallery"),
		GeneratedImageID: record.ID,
		UserID:           user.ID,
		Title:            title,
		Description:      strings.TrimSpace(input.Description),
		Tags:             normalizeTags(input.Tags),
		ImageURL:         record.ImageURL,
		Width:            record.Width,
		Height:           record.Height,
		MimeType:         record.MimeType,
		Model:            record.Model,
		Prompt:           record.Prompt,
		Source:           record.Source,
		ShowPrompt:       input.ShowPrompt,
		Status:           model.GalleryStatusPublic,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	item, err = repository.SaveGalleryImage(item)
	if err != nil {
		return item, err
	}
	record.IsPublished = true
	record.UpdatedAt = now
	_, err = repository.SaveGeneratedImageRecord(record)
	return item, err
}

func ListGalleryImages(q model.Query) (model.GalleryImageList, error) {
	items, total, err := repository.ListGalleryImages(q, false)
	if err != nil {
		return model.GalleryImageList{}, err
	}
	tags, err := repository.ListGalleryTags(q, false)
	if err != nil {
		return model.GalleryImageList{}, err
	}
	return model.GalleryImageList{Items: publicGalleryImages(items), Tags: tags, Total: int(total)}, nil
}

func ListAdminGalleryImages(q model.Query) (model.GalleryImageList, error) {
	items, total, err := repository.ListGalleryImages(q, true)
	if err != nil {
		return model.GalleryImageList{}, err
	}
	tags, err := repository.ListGalleryTags(q, true)
	if err != nil {
		return model.GalleryImageList{}, err
	}
	return model.GalleryImageList{Items: items, Tags: tags, Total: int(total)}, nil
}

func UpdateAdminGalleryImage(id string, input UpdateGalleryImageInput) (model.GalleryImage, error) {
	item, ok, err := repository.GetGalleryImageByID(id)
	if err != nil || !ok {
		if err != nil {
			return item, err
		}
		return item, safeMessageError{message: "画廊作品不存在"}
	}
	item.Title = firstNonEmpty(strings.TrimSpace(input.Title), item.Title)
	item.Description = strings.TrimSpace(input.Description)
	item.Tags = normalizeTags(input.Tags)
	item.ShowPrompt = input.ShowPrompt
	if input.Status != "" {
		if !isValidGalleryStatus(input.Status) {
			return item, safeMessageError{message: "画廊状态无效"}
		}
		item.Status = input.Status
	}
	item.Recommended = input.Recommended
	item.UpdatedAt = now()
	return repository.SaveGalleryImage(item)
}

func AdminSetGalleryStatus(id string, status model.GalleryStatus) (model.GalleryImage, error) {
	if !isValidGalleryStatus(status) {
		return model.GalleryImage{}, safeMessageError{message: "画廊状态无效"}
	}
	item, ok, err := repository.GetGalleryImageByID(id)
	if err != nil || !ok {
		if err != nil {
			return item, err
		}
		return item, safeMessageError{message: "画廊作品不存在"}
	}
	item.Status = status
	item.UpdatedAt = now()
	item, err = repository.SaveGalleryImage(item)
	if err == nil && status == model.GalleryStatusDeleted {
		record, ok, recordErr := repository.GetGeneratedImageRecordByID(item.GeneratedImageID)
		if recordErr == nil && ok {
			record.IsPublished = false
			record.UpdatedAt = now()
			_, recordErr = repository.SaveGeneratedImageRecord(record)
		}
		err = recordErr
	}
	return item, err
}

func publicGalleryImages(items []model.GalleryImage) []model.GalleryImage {
	for i := range items {
		if !items[i].ShowPrompt {
			items[i].Prompt = ""
		}
	}
	return items
}

func normalizeGeneratedImageSource(source string) string {
	source = strings.TrimSpace(source)
	if source == "" {
		return "image-page"
	}
	return source
}

func isAllowedGeneratedImageSource(source string) bool {
	switch source {
	case "image-page", "canvas-node", "canvas-edit":
		return true
	default:
		return false
	}
}

func isValidGalleryStatus(status model.GalleryStatus) bool {
	switch status {
	case model.GalleryStatusPublic, model.GalleryStatusHidden, model.GalleryStatusDeleted:
		return true
	default:
		return false
	}
}

func normalizeTags(values []string) []string {
	seen := map[string]bool{}
	tags := []string{}
	for _, value := range values {
		tag := strings.TrimSpace(value)
		if tag == "" || seen[tag] {
			continue
		}
		seen[tag] = true
		tags = append(tags, tag)
	}
	return tags
}

func truncateRunes(value string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}
