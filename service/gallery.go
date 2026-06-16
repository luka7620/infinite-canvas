package service

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/basketikun/infinite-canvas/model"
	"github.com/basketikun/infinite-canvas/repository"
	"gorm.io/gorm"
)

const maxGalleryImageFileBytes int64 = 32 << 20

var galleryImageHTTPClient = http.Client{Timeout: 30 * time.Second}

type downloadedGalleryImage struct {
	Data     []byte
	MimeType string
	Width    int
	Height   int
}

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

type CreateGalleryCommentInput struct {
	Content string `json:"content"`
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

func GetGeneratedImageFile(id string, user model.AuthUser) ([]byte, string, error) {
	record, ok, err := repository.GetGeneratedImageRecordByID(strings.TrimSpace(id))
	if err != nil || !ok {
		if err != nil {
			return nil, "", err
		}
		return nil, "", safeMessageError{message: "生成图片记录不存在"}
	}
	if record.UserID != user.ID {
		return nil, "", safeMessageError{message: "只能读取自己的生成图片"}
	}
	downloaded, err := readGeneratedImageFile(record.ImageURL)
	if err != nil {
		return nil, "", err
	}
	return downloaded.Data, downloaded.MimeType, nil
}

func PublishGalleryImage(user model.AuthUser, input PublishGalleryImageInput) (model.GalleryImage, error) {
	interactionSetting, err := GalleryInteractionSetting()
	if err != nil {
		return model.GalleryImage{}, err
	}
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
	if !isCloudImageURL(record.ImageURL) || !isGPTImageModel(record.Model) {
		return model.GalleryImage{}, safeMessageError{message: "只有使用云端链接的 GPT 模型生成图片可以上传到画廊"}
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
	downloaded, err := downloadGalleryImage(record.ImageURL)
	if err != nil {
		return model.GalleryImage{}, err
	}
	now := now()
	galleryID := newID("gallery")
	if record.Width <= 0 {
		record.Width = downloaded.Width
	}
	if record.Height <= 0 {
		record.Height = downloaded.Height
	}
	if record.MimeType == "" {
		record.MimeType = downloaded.MimeType
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "未命名作品"
		if input.ShowPrompt {
			title = firstNonEmpty(strings.TrimSpace(record.Prompt), title)
		}
	}
	title = truncateRunes(title, 60)
	prompt := ""
	if input.ShowPrompt {
		prompt = record.Prompt
	}
	item := model.GalleryImage{
		ID:               galleryID,
		GeneratedImageID: record.ID,
		UserID:           user.ID,
		Username:         user.Username,
		DisplayName:      user.DisplayName,
		AvatarURL:        user.AvatarURL,
		Title:            title,
		Description:      strings.TrimSpace(input.Description),
		Tags:             normalizeTags(input.Tags),
		ImageURL:         galleryImageURL(galleryID),
		Width:            record.Width,
		Height:           record.Height,
		MimeType:         record.MimeType,
		Model:            record.Model,
		Prompt:           prompt,
		Source:           record.Source,
		ShowPrompt:       input.ShowPrompt,
		Status:           model.GalleryStatusPublic,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	file := model.GalleryImageFile{
		ID:        newID("gfile"),
		GalleryID: galleryID,
		SourceURL: record.ImageURL,
		MimeType:  downloaded.MimeType,
		Size:      int64(len(downloaded.Data)),
		Data:      downloaded.Data,
		CreatedAt: now,
		UpdatedAt: now,
	}
	record.IsPublished = true
	record.UpdatedAt = now
	item, err = repository.CreateGalleryImageWithFile(item, file, record)
	if err == nil {
		rewardCredits, rewardErr := saveGalleryInteractionReward(user.ID, model.CreditLogTypeGalleryPublishReward, item.ID, interactionSetting.UploadRewardCredits, interactionSetting.DailyUploadLimit, "上传画廊奖励")
		if rewardErr != nil {
			log.Printf("gallery publish reward failed: %v", rewardErr)
		} else {
			item.RewardCredits = rewardCredits
		}
	}
	return item, err
}

func ListGalleryImages(q model.Query, userID string) (model.GalleryImageList, error) {
	items, total, err := repository.ListGalleryImages(q, false)
	if err != nil {
		return model.GalleryImageList{}, err
	}
	items, err = repository.MarkGalleryLiked(items, userID)
	if err != nil {
		return model.GalleryImageList{}, err
	}
	tags, err := repository.ListGalleryTags(q, false)
	if err != nil {
		return model.GalleryImageList{}, err
	}
	return model.GalleryImageList{Items: publicGalleryImages(items), Tags: tags, Total: int(total)}, nil
}

func ListMyLikedGalleryImages(user model.AuthUser, q model.Query) (model.GalleryImageList, error) {
	if strings.TrimSpace(user.ID) == "" || user.Role == model.UserRoleGuest {
		return model.GalleryImageList{}, safeMessageError{message: "请先登录"}
	}
	items, total, err := repository.ListLikedGalleryImages(user.ID, q)
	if err != nil {
		return model.GalleryImageList{}, err
	}
	items, err = repository.MarkGalleryLiked(items, user.ID)
	if err != nil {
		return model.GalleryImageList{}, err
	}
	return model.GalleryImageList{Items: publicGalleryImages(items), Tags: []string{}, Total: int(total)}, nil
}

func ListMyReceivedLikeGalleryImages(user model.AuthUser, q model.Query) (model.GalleryImageList, error) {
	if strings.TrimSpace(user.ID) == "" || user.Role == model.UserRoleGuest {
		return model.GalleryImageList{}, safeMessageError{message: "请先登录"}
	}
	items, total, err := repository.ListReceivedLikeGalleryImages(user.ID, q)
	if err != nil {
		return model.GalleryImageList{}, err
	}
	items, err = repository.MarkGalleryLiked(items, user.ID)
	if err != nil {
		return model.GalleryImageList{}, err
	}
	return model.GalleryImageList{Items: publicGalleryImages(items), Tags: []string{}, Total: int(total)}, nil
}

func MyGalleryRewardStats(user model.AuthUser) (model.GalleryRewardStats, error) {
	if strings.TrimSpace(user.ID) == "" || user.Role == model.UserRoleGuest {
		return model.GalleryRewardStats{}, safeMessageError{message: "请先登录"}
	}
	start, end := galleryInteractionDayRange()
	stats, err := repository.CreditLogTypeStats(user.ID, []model.CreditLogType{
		model.CreditLogTypeGalleryPublishReward,
		model.CreditLogTypeGalleryLikeReward,
		model.CreditLogTypeGalleryLikeAuthorReward,
	}, start, end)
	if err != nil {
		return model.GalleryRewardStats{}, err
	}
	return model.GalleryRewardStats{
		Date:                      checkInDate(),
		UploadRewardCount:         stats[model.CreditLogTypeGalleryPublishReward].Count,
		LikeRewardCount:           stats[model.CreditLogTypeGalleryLikeReward].Count,
		ReceivedLikeRewardCount:   stats[model.CreditLogTypeGalleryLikeAuthorReward].Count,
		ReceivedLikeRewardCredits: stats[model.CreditLogTypeGalleryLikeAuthorReward].Amount,
	}, nil
}

func ToggleGalleryLike(id string, user model.AuthUser) (model.GalleryLikeResult, error) {
	if strings.TrimSpace(user.ID) == "" || user.Role == model.UserRoleGuest {
		return model.GalleryLikeResult{}, safeMessageError{message: "请先登录"}
	}
	galleryID := strings.TrimSpace(id)
	alreadyLiked, err := repository.HasGalleryLike(galleryID, user.ID)
	if err != nil {
		return model.GalleryLikeResult{}, err
	}
	interactionSetting := model.GalleryInteractionSetting{}
	if !alreadyLiked {
		interactionSetting, err = GalleryInteractionSetting()
		if err != nil {
			return model.GalleryLikeResult{}, err
		}
	}
	item, liked, err := repository.ToggleGalleryLike(galleryID, user.ID, now())
	if err != nil {
		return model.GalleryLikeResult{}, galleryNotFoundError(err)
	}
	if liked {
		rewardCredits, rewardErr := saveGalleryInteractionReward(user.ID, model.CreditLogTypeGalleryLikeReward, item.ID, interactionSetting.LikeRewardCredits, interactionSetting.DailyLikeLimit, "点赞画廊奖励")
		if rewardErr != nil {
			log.Printf("gallery like reward failed: %v", rewardErr)
		} else {
			item.RewardCredits = rewardCredits
		}
		if item.UserID != "" && item.UserID != user.ID {
			_, authorRewardErr := saveGalleryInteractionReward(
				item.UserID,
				model.CreditLogTypeGalleryLikeAuthorReward,
				galleryAuthorLikeRewardID(item.ID, user.ID),
				interactionSetting.ReceivedLikeRewardCredits,
				interactionSetting.DailyReceivedLikeLimit,
				"作品被点赞奖励",
			)
			if authorRewardErr != nil {
				log.Printf("gallery like author reward failed: %v", authorRewardErr)
			}
		}
	}
	return model.GalleryLikeResult{Image: publicGalleryImage(item), Liked: liked, RewardCredits: item.RewardCredits}, nil
}

func ListGalleryComments(id string, q model.Query) (model.GalleryCommentList, error) {
	if _, ok, err := repository.GetPublicGalleryImageByID(strings.TrimSpace(id)); err != nil || !ok {
		if err != nil {
			return model.GalleryCommentList{}, err
		}
		return model.GalleryCommentList{}, safeMessageError{message: "画廊作品不存在"}
	}
	items, total, err := repository.ListGalleryComments(strings.TrimSpace(id), q)
	if err != nil {
		return model.GalleryCommentList{}, err
	}
	return model.GalleryCommentList{Items: items, Total: int(total)}, nil
}

func CreateGalleryComment(id string, user model.AuthUser, input CreateGalleryCommentInput) (model.GalleryComment, error) {
	if strings.TrimSpace(user.ID) == "" || user.Role == model.UserRoleGuest {
		return model.GalleryComment{}, safeMessageError{message: "请先登录"}
	}
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return model.GalleryComment{}, safeMessageError{message: "请输入评论内容"}
	}
	if len([]rune(content)) > 500 {
		return model.GalleryComment{}, safeMessageError{message: "评论最多 500 字"}
	}
	now := now()
	comment := model.GalleryComment{
		ID:          newID("gcomment"),
		GalleryID:   strings.TrimSpace(id),
		UserID:      user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Content:     content,
		Status:      "public",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	comment, err := repository.SaveGalleryComment(comment, now)
	if err != nil {
		return model.GalleryComment{}, galleryNotFoundError(err)
	}
	return comment, nil
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
		if !isAdminEditableGalleryStatus(input.Status) {
			return item, safeMessageError{message: "画廊状态无效"}
		}
		item.Status = input.Status
	}
	item.Recommended = input.Recommended
	item.UpdatedAt = now()
	return repository.SaveGalleryImage(item)
}

func AdminSetGalleryStatus(id string, status model.GalleryStatus) (model.GalleryImage, error) {
	if !isAdminEditableGalleryStatus(status) {
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
	return repository.SaveGalleryImage(item)
}

func DeleteAdminGalleryImage(id string) error {
	item, ok, err := repository.GetGalleryImageByID(id)
	if err != nil || !ok {
		if err != nil {
			return err
		}
		return safeMessageError{message: "画廊作品不存在"}
	}
	return repository.DeleteGalleryImage(item.ID, item.GeneratedImageID, now())
}

func GetGalleryImageFile(id string, admin bool) (model.GalleryImageFile, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.GalleryImageFile{}, safeMessageError{message: "画廊作品不存在"}
	}
	var ok bool
	var err error
	if admin {
		_, ok, err = repository.GetGalleryImageByID(id)
	} else {
		_, ok, err = repository.GetPublicGalleryImageByID(id)
	}
	if err != nil || !ok {
		if err != nil {
			return model.GalleryImageFile{}, err
		}
		return model.GalleryImageFile{}, safeMessageError{message: "画廊作品不存在"}
	}
	file, ok, err := repository.GetGalleryImageFileByGalleryID(id)
	if err != nil {
		return model.GalleryImageFile{}, err
	}
	if !ok {
		return model.GalleryImageFile{}, safeMessageError{message: "画廊图片不存在"}
	}
	return file, nil
}

func galleryInteractionDayRange() (string, string) {
	date := checkInDate()
	parsed, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return date + "T00:00:00", ""
	}
	return date + "T00:00:00", parsed.AddDate(0, 0, 1).Format(time.DateOnly) + "T00:00:00"
}

func galleryAuthorLikeRewardID(galleryID string, likerID string) string {
	return galleryID + ":" + likerID
}

func saveGalleryInteractionReward(userID string, logType model.CreditLogType, relatedID string, credits int, dailyLimit int, remark string) (int, error) {
	if strings.TrimSpace(userID) == "" {
		return 0, nil
	}
	if credits < 0 {
		credits = 0
	}
	rewarded, err := repository.HasCreditLog(userID, logType, relatedID)
	if err != nil {
		return 0, err
	}
	if rewarded {
		return 0, nil
	}
	if dailyLimit > 0 {
		start, end := galleryInteractionDayRange()
		total, err := repository.CountCreditLogsByType(userID, logType, start, end)
		if err != nil {
			return 0, err
		}
		if int(total) >= dailyLimit {
			return 0, nil
		}
	}
	changedAt := now()
	_, ok, err := repository.AddUserCreditsWithLog(userID, credits, changedAt, model.CreditLog{
		ID:        newID("credit"),
		Type:      logType,
		RelatedID: relatedID,
		Remark:    remark,
		CreatedAt: changedAt,
	})
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, safeMessageError{message: "用户不存在"}
	}
	return credits, nil
}

func publicGalleryImages(items []model.GalleryImage) []model.GalleryImage {
	for i := range items {
		items[i] = publicGalleryImage(items[i])
	}
	return items
}

func publicGalleryImage(item model.GalleryImage) model.GalleryImage {
	if !item.ShowPrompt {
		item.Prompt = ""
	}
	return item
}

func galleryNotFoundError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return safeMessageError{message: "画廊作品不存在"}
	}
	return err
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

func isCloudImageURL(value string) bool {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return false
	}
	return parsed.Host != "" && (parsed.Scheme == "http" || parsed.Scheme == "https")
}

func galleryImageURL(id string) string {
	return "/api/gallery/" + url.PathEscape(id) + "/image"
}

func readGeneratedImageFile(imageURL string) (downloadedGalleryImage, error) {
	imageURL = strings.TrimSpace(imageURL)
	if strings.HasPrefix(imageURL, "data:") {
		return decodeGeneratedImageDataURL(imageURL)
	}
	return downloadGalleryImage(imageURL)
}

func decodeGeneratedImageDataURL(value string) (downloadedGalleryImage, error) {
	header, content, ok := strings.Cut(value, ",")
	if !ok || !strings.Contains(header, ";base64") {
		return downloadedGalleryImage{}, safeMessageError{message: "生成图片读取失败"}
	}
	data, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return downloadedGalleryImage{}, safeMessageError{message: "生成图片读取失败"}
	}
	if int64(len(data)) > maxGalleryImageFileBytes {
		return downloadedGalleryImage{}, safeMessageError{message: "生成图片不能超过 32MB"}
	}
	mimeType := strings.TrimPrefix(strings.Split(header, ";")[0], "data:")
	mimeType = galleryImageMimeType(mimeType, data)
	if len(data) == 0 || !isAllowedGalleryImageMimeType(mimeType) {
		return downloadedGalleryImage{}, safeMessageError{message: "生成图片文件无效"}
	}
	width, height := galleryImageSize(data)
	return downloadedGalleryImage{Data: data, MimeType: mimeType, Width: width, Height: height}, nil
}

func downloadGalleryImage(imageURL string) (downloadedGalleryImage, error) {
	request, err := http.NewRequest(http.MethodGet, strings.TrimSpace(imageURL), nil)
	if err != nil {
		return downloadedGalleryImage{}, safeMessageError{message: "画廊图片保存失败"}
	}
	response, err := galleryImageHTTPClient.Do(request)
	if err != nil {
		return downloadedGalleryImage{}, safeMessageError{message: "画廊图片保存失败"}
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusBadRequest {
		return downloadedGalleryImage{}, safeMessageError{message: "画廊图片保存失败"}
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxGalleryImageFileBytes+1))
	if err != nil {
		return downloadedGalleryImage{}, err
	}
	if int64(len(data)) > maxGalleryImageFileBytes {
		return downloadedGalleryImage{}, safeMessageError{message: "画廊图片不能超过 32MB"}
	}
	mimeType := galleryImageMimeType(response.Header.Get("Content-Type"), data)
	if len(data) == 0 || !isAllowedGalleryImageMimeType(mimeType) {
		return downloadedGalleryImage{}, safeMessageError{message: "画廊图片文件无效"}
	}
	width, height := galleryImageSize(data)
	return downloadedGalleryImage{Data: data, MimeType: mimeType, Width: width, Height: height}, nil
}

func galleryImageMimeType(contentType string, data []byte) string {
	if mediaType, _, err := mime.ParseMediaType(contentType); err == nil && strings.HasPrefix(strings.ToLower(mediaType), "image/") {
		return strings.ToLower(mediaType)
	}
	if detected := strings.ToLower(http.DetectContentType(data)); strings.HasPrefix(detected, "image/") {
		return detected
	}
	if _, format, err := image.DecodeConfig(bytes.NewReader(data)); err == nil {
		return galleryImageFormatMimeType(format)
	}
	return ""
}

func galleryImageSize(data []byte) (int, int) {
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0
	}
	return config.Width, config.Height
}

func galleryImageFormatMimeType(format string) string {
	switch strings.ToLower(format) {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	default:
		return "image/" + strings.ToLower(format)
	}
}

func isAllowedGalleryImageMimeType(mimeType string) bool {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/png", "image/jpeg", "image/gif", "image/webp", "image/avif":
		return true
	default:
		return false
	}
}

func isGPTImageModel(modelName string) bool {
	name := strings.ToLower(strings.TrimSpace(modelName))
	if index := strings.LastIndex(name, "/"); index >= 0 {
		name = name[index+1:]
	}
	return strings.HasPrefix(name, "gpt-") || strings.HasPrefix(name, "gpt_") || strings.HasPrefix(name, "gpt4") || strings.HasPrefix(name, "gpt5")
}

func isAdminEditableGalleryStatus(status model.GalleryStatus) bool {
	switch status {
	case model.GalleryStatusPublic, model.GalleryStatusHidden:
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
