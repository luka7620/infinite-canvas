package model

type GalleryStatus string

const (
	GalleryStatusPublic  GalleryStatus = "public"
	GalleryStatusHidden  GalleryStatus = "hidden"
	GalleryStatusDeleted GalleryStatus = "deleted"
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
	CreatedAt        string        `json:"createdAt"`
	UpdatedAt        string        `json:"updatedAt"`
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

