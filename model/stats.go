package model

type AdminStatsRange string

const (
	AdminStatsRange7D  AdminStatsRange = "7d"
	AdminStatsRange30D AdminStatsRange = "30d"
	AdminStatsRangeAll AdminStatsRange = "all"
)

type AdminStats struct {
	Range          AdminStatsRange         `json:"range"`
	StartDate      string                  `json:"startDate"`
	EndDate        string                  `json:"endDate"`
	TrendStartDate string                  `json:"trendStartDate"`
	TrendEndDate   string                  `json:"trendEndDate"`
	Summary        AdminStatsSummary       `json:"summary"`
	Trends         []AdminStatsTrendDay    `json:"trends"`
	Distributions  AdminStatsDistributions `json:"distributions"`
}

type AdminStatsSummary struct {
	UserTotal           int `json:"userTotal"`
	NewUsers            int `json:"newUsers"`
	CreditBalanceTotal  int `json:"creditBalanceTotal"`
	CreditConsumed      int `json:"creditConsumed"`
	CreditRefunded      int `json:"creditRefunded"`
	GeneratedImages     int `json:"generatedImages"`
	PublicGalleryImages int `json:"publicGalleryImages"`
	PromptTotal         int `json:"promptTotal"`
	AssetTotal          int `json:"assetTotal"`
	InviteCodeTotal     int `json:"inviteCodeTotal"`
	InviteCodeEnabled   int `json:"inviteCodeEnabled"`
}

type AdminStatsTrendDay struct {
	Date             string `json:"date"`
	NewUsers         int    `json:"newUsers"`
	CreditConsumed   int    `json:"creditConsumed"`
	CreditRefunded   int    `json:"creditRefunded"`
	GeneratedImages  int    `json:"generatedImages"`
	GalleryPublishes int    `json:"galleryPublishes"`
}

type AdminStatsDistributions struct {
	UserStatus    []AdminStatsDistributionItem `json:"userStatus"`
	LoginSource   []AdminStatsDistributionItem `json:"loginSource"`
	CreditLogType []AdminStatsDistributionItem `json:"creditLogType"`
	GalleryStatus []AdminStatsDistributionItem `json:"galleryStatus"`
	ModelUsage    []AdminStatsDistributionItem `json:"modelUsage"`
	AssetType     []AdminStatsDistributionItem `json:"assetType"`
}

type AdminStatsDistributionItem struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value int    `json:"value"`
}

type AdminStatsDailyCount struct {
	Date  string `json:"date"`
	Value int    `json:"value"`
}

type AdminStatsDailyCredit struct {
	Date  string        `json:"date"`
	Type  CreditLogType `json:"type"`
	Value int           `json:"value"`
}
