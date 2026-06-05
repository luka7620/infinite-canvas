package repository

import (
	"encoding/json"

	"github.com/basketikun/infinite-canvas/model"
	"gorm.io/gorm/clause"
)

// GetSettings 返回 public 和 private 两行配置。
func GetSettings() (model.Settings, error) {
	if client := Redis(); client != nil {
		var cached model.Settings
		if data, err := client.Get(cacheContext(), CacheKey("settings")).Bytes(); err == nil {
			if json.Unmarshal(data, &cached) == nil {
				return cached, nil
			}
		}
	}
	db, err := DB()
	if err != nil {
		return model.Settings{}, err
	}
	var items []model.Setting
	if err := db.Find(&items).Error; err != nil {
		return model.Settings{}, err
	}
	result := model.Settings{}
	for _, item := range items {
		if item.Key == model.SettingKeyPrivate {
			_ = json.Unmarshal(item.Value, &result.Private)
		} else if item.Key == model.SettingKeyPublic {
			_ = json.Unmarshal(item.Value, &result.Public)
		}
	}
	if client := Redis(); client != nil {
		if data, err := json.Marshal(result); err == nil {
			_ = client.Set(cacheContext(), CacheKey("settings"), data, CacheTTL()).Err()
		}
	}
	return result, nil
}

// SaveSettings 保存 public 和 private 两行配置。
func SaveSettings(settings model.Settings, now string) (model.Settings, error) {
	db, err := DB()
	if err != nil {
		return settings, err
	}
	publicValue, _ := json.Marshal(settings.Public)
	privateValue, _ := json.Marshal(settings.Private)
	items := []model.Setting{
		{Key: model.SettingKeyPublic, Value: publicValue, CreatedAt: now, UpdatedAt: now},
		{Key: model.SettingKeyPrivate, Value: privateValue, CreatedAt: now, UpdatedAt: now},
	}
	err = db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&items).Error
	if err == nil {
		if client := Redis(); client != nil {
			_ = client.Del(cacheContext(), CacheKey("settings")).Err()
		}
	}
	return settings, err
}
