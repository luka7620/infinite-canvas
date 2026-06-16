package model

type UserRole string

const (
	UserRoleGuest UserRole = "guest"
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

type UserStatus string

const (
	UserStatusActive UserStatus = "active"
	UserStatusBan    UserStatus = "ban"
)

type UserIdentityProvider string

const (
	UserIdentityProviderLinuxDo UserIdentityProvider = "linux-do"
)

// User 系统用户。
type User struct {
	ID                string     `json:"id" gorm:"primaryKey"`
	Username          string     `json:"username" gorm:"uniqueIndex"`
	Password          string     `json:"password,omitempty"`
	Email             string     `json:"email"`
	DisplayName       string     `json:"displayName"`
	AvatarURL         string     `json:"avatarUrl"`
	Role              UserRole   `json:"role"`
	Credits           int        `json:"credits"`
	AffCode           string     `json:"affCode" gorm:"uniqueIndex"`
	AffCount          int        `json:"affCount"`
	InviterID         string     `json:"inviterId"`
	GithubID          string     `json:"githubId"`
	LinuxDoID         string     `json:"linuxDoId" gorm:"index"`
	WechatID          string     `json:"wechatId"`
	Status            UserStatus `json:"status"`
	LastLoginAt       string     `json:"lastLoginAt"`
	LastCheckInDate   string     `json:"lastCheckInDate"`
	LikeGivenCount    int        `json:"likeGivenCount" gorm:"-"`
	LikeReceivedCount int        `json:"likeReceivedCount" gorm:"-"`
	Extra             string     `json:"extra" gorm:"type:text"`
	CreatedAt         string     `json:"createdAt"`
	UpdatedAt         string     `json:"updatedAt"`
}

type UserLikeStats struct {
	Given    int `json:"given"`
	Received int `json:"received"`
}

type UserIdentity struct {
	ID         string               `json:"id" gorm:"primaryKey"`
	UserID     string               `json:"userId" gorm:"index"`
	Provider   UserIdentityProvider `json:"provider" gorm:"index;uniqueIndex:idx_user_identity_provider_external"`
	ExternalID string               `json:"externalId" gorm:"index;uniqueIndex:idx_user_identity_provider_external"`
	CreatedAt  string               `json:"createdAt"`
}

// UserList 用户分页结果。
type UserList struct {
	Items []User `json:"items"`
	Total int    `json:"total"`
}

// AuthUser 用户公开信息。
type AuthUser struct {
	ID                string   `json:"id"`
	Username          string   `json:"username"`
	DisplayName       string   `json:"displayName"`
	AvatarURL         string   `json:"avatarUrl"`
	Role              UserRole `json:"role"`
	Credits           int      `json:"credits"`
	LastCheckInDate   string   `json:"lastCheckInDate"`
	CheckedInToday    bool     `json:"checkedInToday"`
	LikeGivenCount    int      `json:"likeGivenCount"`
	LikeReceivedCount int      `json:"likeReceivedCount"`
	CreatedAt         string   `json:"createdAt"`
	UpdatedAt         string   `json:"updatedAt"`
}

// AuthSession 登录会话信息。
type AuthSession struct {
	Token string   `json:"token"`
	User  AuthUser `json:"user"`
}

func PublicUser(user User) AuthUser {
	return AuthUser{
		ID:              user.ID,
		Username:        user.Username,
		DisplayName:     user.DisplayName,
		AvatarURL:       user.AvatarURL,
		Role:            user.Role,
		Credits:         user.Credits,
		LastCheckInDate: user.LastCheckInDate,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}

type CheckInResult struct {
	User    AuthUser `json:"user"`
	Credits int      `json:"credits"`
}

type CreditLogType string

const (
	CreditLogTypeAdminAdjust          CreditLogType = "admin_adjust"
	CreditLogTypeRegisterBonus        CreditLogType = "register_bonus"
	CreditLogTypeAIConsume            CreditLogType = "ai_consume"
	CreditLogTypeAIRefund             CreditLogType = "ai_refund"
	CreditLogTypeCheckIn              CreditLogType = "check_in"
	CreditLogTypeInviteCode           CreditLogType = "invite_code"
	CreditLogTypeGalleryPublishReward    CreditLogType = "gallery_publish_reward"
	CreditLogTypeGalleryLikeReward       CreditLogType = "gallery_like_reward"
	CreditLogTypeGalleryLikeAuthorReward CreditLogType = "gallery_like_author_reward"
)

// CreditLog 用户算力点变更流水。
type CreditLog struct {
	ID        string        `json:"id" gorm:"primaryKey"`
	UserID    string        `json:"userId" gorm:"index"`
	Type      CreditLogType `json:"type"`
	Amount    int           `json:"amount"`
	Balance   int           `json:"balance"`
	RelatedID string        `json:"relatedId"`
	Remark    string        `json:"remark"`
	Extra     string        `json:"extra" gorm:"type:text"`
	CreatedAt string        `json:"createdAt"`
}

type AdminCreditLog struct {
	CreditLog
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}

type CreditLogList struct {
	Items []AdminCreditLog `json:"items"`
	Total int              `json:"total"`
}
