package repository

import (
	"sort"
	"time"

	"github.com/basketikun/infinite-canvas/model"
	"gorm.io/gorm"
)

type statsValueRow struct {
	Value int64 `gorm:"column:value"`
}

type statsStringRow struct {
	Value string `gorm:"column:value"`
}

type statsCountRow struct {
	StatsKey string `gorm:"column:stats_key"`
	Total    int64  `gorm:"column:total"`
}

type statsDateCountRow struct {
	StatsDate string `gorm:"column:stats_date"`
	Total     int64  `gorm:"column:total"`
}

type statsDailyCreditRow struct {
	StatsDate string              `gorm:"column:stats_date"`
	Type      model.CreditLogType `gorm:"column:type"`
	Total     int64               `gorm:"column:total"`
}

func AdminStatsSummary(startDate string, endDate string) (model.AdminStatsSummary, error) {
	db, err := DB()
	if err != nil {
		return model.AdminStatsSummary{}, err
	}
	var summary model.AdminStatsSummary
	if summary.UserTotal, err = countStats(db.Model(&model.User{})); err != nil {
		return summary, err
	}
	if summary.NewUsers, err = countStats(applyStatsDateRange(db.Model(&model.User{}), startDate, endDate)); err != nil {
		return summary, err
	}
	if summary.CreditBalanceTotal, err = sumStats(db.Model(&model.User{}), "COALESCE(SUM(credits), 0)"); err != nil {
		return summary, err
	}
	if summary.CreditConsumed, err = sumStats(applyStatsDateRange(db.Model(&model.CreditLog{}).Where("type = ?", model.CreditLogTypeAIConsume), startDate, endDate), "COALESCE(SUM(amount), 0)"); err != nil {
		return summary, err
	}
	if summary.CreditConsumed < 0 {
		summary.CreditConsumed = -summary.CreditConsumed
	}
	if summary.CreditRefunded, err = sumStats(applyStatsDateRange(db.Model(&model.CreditLog{}).Where("type = ?", model.CreditLogTypeAIRefund), startDate, endDate), "COALESCE(SUM(amount), 0)"); err != nil {
		return summary, err
	}
	if summary.GeneratedImages, err = countStats(applyStatsDateRange(db.Model(&model.GeneratedImageRecord{}), startDate, endDate)); err != nil {
		return summary, err
	}
	if summary.PublicGalleryImages, err = countStats(db.Model(&model.GalleryImage{}).Where("status = ?", model.GalleryStatusPublic)); err != nil {
		return summary, err
	}
	if summary.PromptTotal, err = countStats(db.Model(&model.Prompt{})); err != nil {
		return summary, err
	}
	if summary.AssetTotal, err = countStats(db.Model(&model.Asset{})); err != nil {
		return summary, err
	}
	if summary.InviteCodeTotal, err = countStats(db.Model(&model.InviteCode{})); err != nil {
		return summary, err
	}
	summary.InviteCodeEnabled, err = countStats(db.Model(&model.InviteCode{}).Where("enabled = ?", true))
	return summary, err
}

func AdminStatsFirstCreatedAt() (string, error) {
	db, err := DB()
	if err != nil {
		return "", err
	}
	values := []string{}
	for _, item := range []any{&model.User{}, &model.CreditLog{}, &model.InviteCode{}, &model.GeneratedImageRecord{}, &model.GalleryImage{}, &model.Prompt{}, &model.Asset{}} {
		value, err := minCreatedAt(db.Model(item))
		if err != nil {
			return "", err
		}
		if value != "" {
			values = append(values, value)
		}
	}
	sort.Strings(values)
	if len(values) == 0 {
		return "", nil
	}
	return values[0], nil
}

func AdminStatsUserDailyCounts(startDate string, endDate string) ([]model.AdminStatsDailyCount, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	return dailyCounts(applyStatsDateRange(db.Model(&model.User{}), startDate, endDate))
}

func AdminStatsGeneratedImageDailyCounts(startDate string, endDate string) ([]model.AdminStatsDailyCount, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	return dailyCounts(applyStatsDateRange(db.Model(&model.GeneratedImageRecord{}), startDate, endDate))
}

func AdminStatsGalleryDailyCounts(startDate string, endDate string) ([]model.AdminStatsDailyCount, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	return dailyCounts(applyStatsDateRange(db.Model(&model.GalleryImage{}), startDate, endDate))
}

func AdminStatsCreditDailySums(startDate string, endDate string) ([]model.AdminStatsDailyCredit, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var rows []statsDailyCreditRow
	err = applyStatsDateRange(db.Model(&model.CreditLog{}).Where("type IN ?", []model.CreditLogType{model.CreditLogTypeAIConsume, model.CreditLogTypeAIRefund}), startDate, endDate).
		Select("SUBSTR(created_at, 1, 10) AS stats_date, type, COALESCE(SUM(amount), 0) AS total").
		Group("SUBSTR(created_at, 1, 10), type").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]model.AdminStatsDailyCredit, 0, len(rows))
	for _, row := range rows {
		value := int(row.Total)
		if row.Type == model.CreditLogTypeAIConsume && value < 0 {
			value = -value
		}
		items = append(items, model.AdminStatsDailyCredit{Date: row.StatsDate, Type: row.Type, Value: value})
	}
	return items, nil
}

func AdminStatsUserStatusDistribution() ([]model.AdminStatsDistributionItem, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	return groupedCounts(db.Model(&model.User{}), "CASE WHEN status IS NULL OR status = '' THEN 'active' ELSE status END")
}

func AdminStatsLoginSourceDistribution() ([]model.AdminStatsDistributionItem, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	return groupedCounts(db.Model(&model.User{}), "CASE WHEN linux_do_id IS NOT NULL AND linux_do_id <> '' THEN 'linux-do' ELSE 'password' END")
}

func AdminStatsCreditLogTypeDistribution(startDate string, endDate string) ([]model.AdminStatsDistributionItem, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	return groupedCounts(applyStatsDateRange(db.Model(&model.CreditLog{}), startDate, endDate), "CASE WHEN type IS NULL OR type = '' THEN 'unknown' ELSE type END")
}

func AdminStatsGalleryStatusDistribution() ([]model.AdminStatsDistributionItem, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	return groupedCounts(db.Model(&model.GalleryImage{}), "CASE WHEN status IS NULL OR status = '' THEN 'public' ELSE status END")
}

func AdminStatsModelUsageDistribution(startDate string, endDate string) ([]model.AdminStatsDistributionItem, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	return groupedCounts(applyStatsDateRange(db.Model(&model.GeneratedImageRecord{}), startDate, endDate), "CASE WHEN model IS NULL OR model = '' THEN 'unknown' ELSE model END")
}

func AdminStatsAssetTypeDistribution() ([]model.AdminStatsDistributionItem, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	return groupedCounts(db.Model(&model.Asset{}), "CASE WHEN type IS NULL OR type = '' THEN 'unknown' ELSE type END")
}

func countStats(tx *gorm.DB) (int, error) {
	var total int64
	err := tx.Count(&total).Error
	return int(total), err
}

func sumStats(tx *gorm.DB, expression string) (int, error) {
	var row statsValueRow
	err := tx.Select(expression + " AS value").Scan(&row).Error
	return int(row.Value), err
}

func minCreatedAt(tx *gorm.DB) (string, error) {
	var row statsStringRow
	err := tx.Where("created_at IS NOT NULL AND created_at <> ''").Select("MIN(created_at) AS value").Scan(&row).Error
	return row.Value, err
}

func dailyCounts(tx *gorm.DB) ([]model.AdminStatsDailyCount, error) {
	var rows []statsDateCountRow
	err := tx.Select("SUBSTR(created_at, 1, 10) AS stats_date, COUNT(*) AS total").
		Group("SUBSTR(created_at, 1, 10)").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]model.AdminStatsDailyCount, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.AdminStatsDailyCount{Date: row.StatsDate, Value: int(row.Total)})
	}
	return items, nil
}

func groupedCounts(tx *gorm.DB, expression string) ([]model.AdminStatsDistributionItem, error) {
	var rows []statsCountRow
	err := tx.Select(expression+" AS stats_key, COUNT(*) AS total").Group(expression).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	items := make([]model.AdminStatsDistributionItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.AdminStatsDistributionItem{Key: row.StatsKey, Value: int(row.Total)})
	}
	return items, nil
}

func applyStatsDateRange(tx *gorm.DB, startDate string, endDate string) *gorm.DB {
	if startDate != "" {
		tx = tx.Where("created_at >= ?", startDate+"T00:00:00")
	}
	if endDate != "" {
		tx = tx.Where("created_at < ?", nextStatsDate(endDate)+"T00:00:00")
	}
	return tx
}

func nextStatsDate(date string) string {
	parsed, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return date
	}
	return parsed.AddDate(0, 0, 1).Format(time.DateOnly)
}
