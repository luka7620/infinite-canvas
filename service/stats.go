package service

import (
	"sort"
	"strings"
	"time"

	"github.com/basketikun/infinite-canvas/model"
	"github.com/basketikun/infinite-canvas/repository"
)

const adminStatsMaxTrendDays = 90

func AdminStats(rangeValue string) (model.AdminStats, error) {
	statsRange := normalizeAdminStatsRange(rangeValue)
	startDate, endDate, err := adminStatsDateRange(statsRange)
	if err != nil {
		return model.AdminStats{}, err
	}
	summary, err := repository.AdminStatsSummary(startDate, endDate)
	if err != nil {
		return model.AdminStats{}, err
	}
	trendStartDate, trendEndDate := adminStatsTrendRange(startDate, endDate)
	trends, err := adminStatsTrends(trendStartDate, trendEndDate)
	if err != nil {
		return model.AdminStats{}, err
	}
	distributions, err := adminStatsDistributions(startDate, endDate)
	if err != nil {
		return model.AdminStats{}, err
	}
	return model.AdminStats{
		Range:         statsRange,
		StartDate:     startDate,
		EndDate:       endDate,
		TrendStartDate: trendStartDate,
		TrendEndDate:   trendEndDate,
		Summary:       summary,
		Trends:        trends,
		Distributions: distributions,
	}, nil
}

func normalizeAdminStatsRange(value string) model.AdminStatsRange {
	trimmed := model.AdminStatsRange(strings.TrimSpace(value))
	switch trimmed {
	case model.AdminStatsRange7D, model.AdminStatsRangeAll:
		return trimmed
	default:
		return model.AdminStatsRange30D
	}
}

func adminStatsDateRange(statsRange model.AdminStatsRange) (string, string, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -29)
	if statsRange == model.AdminStatsRange7D {
		start = end.AddDate(0, 0, -6)
	} else if statsRange == model.AdminStatsRangeAll {
		first, err := repository.AdminStatsFirstCreatedAt()
		if err != nil {
			return "", "", err
		}
		if first == "" {
			start = end
		} else if len(first) >= len(time.DateOnly) {
			if parsed, err := time.Parse(time.DateOnly, first[:len(time.DateOnly)]); err == nil {
				start = parsed
			}
		}
	}
	if start.After(end) {
		start = end
	}
	return start.Format(time.DateOnly), end.Format(time.DateOnly), nil
}

func adminStatsTrendRange(startDate string, endDate string) (string, string) {
	start, startErr := time.Parse(time.DateOnly, startDate)
	end, endErr := time.Parse(time.DateOnly, endDate)
	if startErr != nil || endErr != nil {
		return startDate, endDate
	}
	limitedStart := end.AddDate(0, 0, -(adminStatsMaxTrendDays - 1))
	if start.Before(limitedStart) {
		start = limitedStart
	}
	return start.Format(time.DateOnly), end.Format(time.DateOnly)
}

func adminStatsTrends(startDate string, endDate string) ([]model.AdminStatsTrendDay, error) {
	userRows, err := repository.AdminStatsUserDailyCounts(startDate, endDate)
	if err != nil {
		return nil, err
	}
	creditRows, err := repository.AdminStatsCreditDailySums(startDate, endDate)
	if err != nil {
		return nil, err
	}
	generatedRows, err := repository.AdminStatsGeneratedImageDailyCounts(startDate, endDate)
	if err != nil {
		return nil, err
	}
	galleryRows, err := repository.AdminStatsGalleryDailyCounts(startDate, endDate)
	if err != nil {
		return nil, err
	}

	byDate := map[string]model.AdminStatsTrendDay{}
	for _, date := range dateRange(startDate, endDate) {
		byDate[date] = model.AdminStatsTrendDay{Date: date}
	}
	for _, row := range userRows {
		item := byDate[row.Date]
		item.Date = row.Date
		item.NewUsers = row.Value
		byDate[row.Date] = item
	}
	for _, row := range creditRows {
		item := byDate[row.Date]
		item.Date = row.Date
		if row.Type == model.CreditLogTypeAIConsume {
			item.CreditConsumed = row.Value
		} else if row.Type == model.CreditLogTypeAIRefund {
			item.CreditRefunded = row.Value
		}
		byDate[row.Date] = item
	}
	for _, row := range generatedRows {
		item := byDate[row.Date]
		item.Date = row.Date
		item.GeneratedImages = row.Value
		byDate[row.Date] = item
	}
	for _, row := range galleryRows {
		item := byDate[row.Date]
		item.Date = row.Date
		item.GalleryPublishes = row.Value
		byDate[row.Date] = item
	}

	trends := make([]model.AdminStatsTrendDay, 0, len(byDate))
	for _, item := range byDate {
		trends = append(trends, item)
	}
	sort.Slice(trends, func(i, j int) bool { return trends[i].Date < trends[j].Date })
	return trends, nil
}

func adminStatsDistributions(startDate string, endDate string) (model.AdminStatsDistributions, error) {
	var result model.AdminStatsDistributions
	var err error
	if result.UserStatus, err = repository.AdminStatsUserStatusDistribution(); err != nil {
		return result, err
	}
	if result.LoginSource, err = repository.AdminStatsLoginSourceDistribution(); err != nil {
		return result, err
	}
	if result.CreditLogType, err = repository.AdminStatsCreditLogTypeDistribution(startDate, endDate); err != nil {
		return result, err
	}
	if result.GalleryStatus, err = repository.AdminStatsGalleryStatusDistribution(); err != nil {
		return result, err
	}
	if result.ModelUsage, err = repository.AdminStatsModelUsageDistribution(startDate, endDate); err != nil {
		return result, err
	}
	if result.AssetType, err = repository.AdminStatsAssetTypeDistribution(); err != nil {
		return result, err
	}
	result.UserStatus = labelStatsItems(result.UserStatus, map[string]string{"active": "正常", "ban": "禁用"})
	result.LoginSource = labelStatsItems(result.LoginSource, map[string]string{"password": "账号密码", "linux-do": "Linux.do"})
	result.CreditLogType = labelStatsItems(result.CreditLogType, map[string]string{
		string(model.CreditLogTypeAdminAdjust):          "后台调整",
		string(model.CreditLogTypeRegisterBonus):        "注册赠送",
		string(model.CreditLogTypeAIConsume):            "模型消费",
		string(model.CreditLogTypeAIRefund):             "失败返还",
		string(model.CreditLogTypeCheckIn):              "每日签到",
		string(model.CreditLogTypeInviteCode):           "邀请码兑换",
		string(model.CreditLogTypeGalleryPublishReward):    "上传画廊奖励",
		string(model.CreditLogTypeGalleryLikeReward):       "点赞画廊奖励",
		string(model.CreditLogTypeGalleryLikeAuthorReward): "作品被点赞奖励",
	})
	result.GalleryStatus = labelStatsItems(result.GalleryStatus, map[string]string{string(model.GalleryStatusPublic): "公开", string(model.GalleryStatusHidden): "隐藏"})
	result.ModelUsage = labelStatsItems(result.ModelUsage, nil)
	result.AssetType = labelStatsItems(result.AssetType, map[string]string{string(model.AssetTypeText): "文本", string(model.AssetTypeImage): "图片", "video": "视频"})
	return result, nil
}

func labelStatsItems(items []model.AdminStatsDistributionItem, labels map[string]string) []model.AdminStatsDistributionItem {
	for i := range items {
		if strings.TrimSpace(items[i].Key) == "" {
			items[i].Key = "unknown"
		}
		if label, ok := labels[items[i].Key]; ok {
			items[i].Label = label
		} else if items[i].Key == "unknown" {
			items[i].Label = "未知"
		} else {
			items[i].Label = items[i].Key
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Value == items[j].Value {
			return items[i].Label < items[j].Label
		}
		return items[i].Value > items[j].Value
	})
	return items
}

func dateRange(startDate string, endDate string) []string {
	start, err := time.Parse(time.DateOnly, startDate)
	if err != nil {
		return []string{}
	}
	end, err := time.Parse(time.DateOnly, endDate)
	if err != nil {
		return []string{}
	}
	dates := []string{}
	for !start.After(end) {
		dates = append(dates, start.Format(time.DateOnly))
		start = start.AddDate(0, 0, 1)
	}
	return dates
}
