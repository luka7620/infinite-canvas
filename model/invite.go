package model

type InviteCodeType string

const (
	InviteCodeTypeRegister InviteCodeType = "register"
	InviteCodeTypeCredits  InviteCodeType = "credits"
)

type InviteCode struct {
	ID        string         `json:"id" gorm:"primaryKey"`
	Code      string         `json:"code" gorm:"uniqueIndex"`
	Type      InviteCodeType `json:"type" gorm:"index"`
	Credits   int            `json:"credits"`
	MaxUses   int            `json:"maxUses"`
	UsedCount int            `json:"usedCount"`
	Enabled   bool           `json:"enabled" gorm:"index"`
	Remark    string         `json:"remark"`
	CreatedAt string         `json:"createdAt"`
	UpdatedAt string         `json:"updatedAt"`
}

type InviteCodeUse struct {
	ID           string         `json:"id" gorm:"primaryKey"`
	InviteCodeID string         `json:"inviteCodeId" gorm:"index;uniqueIndex:idx_invite_code_user"`
	Code         string         `json:"code" gorm:"index"`
	UserID       string         `json:"userId" gorm:"index;uniqueIndex:idx_invite_code_user"`
	Type         InviteCodeType `json:"type" gorm:"index"`
	Credits      int            `json:"credits"`
	CreatedAt    string         `json:"createdAt"`
}

type InviteCodeList struct {
	Items []InviteCode `json:"items"`
	Total int          `json:"total"`
}

type InviteCodeBatchResult struct {
	Items []InviteCode `json:"items"`
}

type InviteCodeRedeemResult struct {
	User       AuthUser   `json:"user"`
	InviteCode InviteCode `json:"inviteCode"`
	Credits    int        `json:"credits"`
}
