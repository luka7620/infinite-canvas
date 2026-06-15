package model

type GalleryStatus string

const (
	GalleryStatusPublic GalleryStatus = "public"
	GalleryStatusHidden GalleryStatus = "hidden"
)

// GeneratedImageRecord 站内模型通道生成的图片记录。
type GeneratedImageRecord struct {
	ID          string `json:"id" gorm:"primaryKey"`
	UserID      string `json:"userId" gorm:"index"`
	ImageURL    string `json:"imageUrl" gorm:"type:text"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	MimeType    string `json:"mimeType"`
	Model       string `json:"model"`
	Prompt      string `json:"prompt" gorm:"type:text"`
	Source      string `json:"source"`
	IsPublished bool   `json:"isPublished" gorm:"index"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// GalleryImage 公开画廊图片。
type GalleryImage struct {
	ID               string        `json:"id" gorm:"primaryKey"`
	GeneratedImageID string        `json:"generatedImageId" gorm:"uniqueIndex"`
	UserID           string        `json:"userId" gorm:"index"`
	Username         string        `json:"username" gorm:"-"`
	DisplayName      string        `json:"displayName" gorm:"-"`
	AvatarURL        string        `json:"avatarUrl" gorm:"-"`
	Title            string        `json:"title"`
	Description      string        `json:"description" gorm:"type:text"`
	Tags             []string      `json:"tags" gorm:"serializer:json"`
	ImageURL         string        `json:"imageUrl" gorm:"type:text"`
	Width            int           `json:"width"`
	Height           int           `json:"height"`
	MimeType         string        `json:"mimeType"`
	Model            string        `json:"model"`
	Prompt           string        `json:"prompt,omitempty" gorm:"type:text"`
	Source           string        `json:"source"`
	ShowPrompt       bool          `json:"showPrompt"`
	Status           GalleryStatus `json:"status" gorm:"index"`
	Recommended      bool          `json:"recommended" gorm:"index"`
	LikeCount        int           `json:"likeCount"`
	CommentCount     int           `json:"commentCount"`
	Liked            bool          `json:"liked" gorm:"-"`
	RewardCredits    int           `json:"rewardCredits,omitempty" gorm:"-"`
	CreatedAt        string        `json:"createdAt"`
	UpdatedAt        string        `json:"updatedAt"`
}

// GalleryImageFile 画廊图片数据库副本。
type GalleryImageFile struct {
	ID        string `json:"id" gorm:"primaryKey"`
	GalleryID string `json:"galleryId" gorm:"uniqueIndex"`
	SourceURL string `json:"sourceUrl" gorm:"type:text"`
	MimeType  string `json:"mimeType"`
	Size      int64  `json:"size"`
	Data      []byte `json:"-"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type GalleryLike struct {
	ID        string `json:"id" gorm:"primaryKey"`
	GalleryID string `json:"galleryId" gorm:"index;uniqueIndex:idx_gallery_like_user"`
	UserID    string `json:"userId" gorm:"index;uniqueIndex:idx_gallery_like_user"`
	CreatedAt string `json:"createdAt"`
}

type GalleryComment struct {
	ID            string `json:"id" gorm:"primaryKey"`
	GalleryID     string `json:"galleryId" gorm:"index"`
	UserID        string `json:"userId" gorm:"index"`
	Username      string `json:"username"`
	DisplayName   string `json:"displayName"`
	AvatarURL     string `json:"avatarUrl" gorm:"type:text"`
	Content       string `json:"content" gorm:"type:text"`
	Status        string `json:"status" gorm:"index"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type GeneratedImageRecordList struct {
	Items []GeneratedImageRecord `json:"items"`
	Total int                    `json:"total"`
}

type GalleryImageList struct {
	Items []GalleryImage `json:"items"`
	Tags  []string       `json:"tags"`
	Total int            `json:"total"`
}

type GalleryCommentList struct {
	Items []GalleryComment `json:"items"`
	Total int              `json:"total"`
}

type GalleryLikeResult struct {
	Image         GalleryImage `json:"image"`
	Liked         bool         `json:"liked"`
	RewardCredits int          `json:"rewardCredits,omitempty"`
}
