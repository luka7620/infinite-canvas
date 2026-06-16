package repository

import (
	"errors"
	"strings"

	"github.com/basketikun/infinite-canvas/model"
	"gorm.io/gorm"
)

// ListUsers 分页查询用户。
func ListUsers(q model.Query) ([]model.User, int64, error) {
	db, err := DB()
	if err != nil {
		return nil, 0, err
	}
	q.Normalize()
	tx := db.Model(&model.User{})
	if keyword := strings.TrimSpace(q.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		tx = tx.Where("username LIKE ? OR display_name LIKE ? OR email LIKE ? OR linux_do_id LIKE ?", like, like, like, like)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []model.User
	err = tx.Order("created_at desc").Offset(q.Offset()).Limit(q.PageSize).Find(&users).Error
	return users, total, err
}

type userLikeCountRow struct {
	UserID string `gorm:"column:user_id"`
	Total  int64  `gorm:"column:total"`
}

type CreditLogTypeStat struct {
	Type   model.CreditLogType
	Count  int
	Amount int
}

type creditLogTypeStatRow struct {
	Type   model.CreditLogType `gorm:"column:type"`
	Count  int64               `gorm:"column:total"`
	Amount int                 `gorm:"column:amount"`
}

type receivedLikeRow struct {
	OwnerID   string `gorm:"column:owner_id"`
	GalleryID string `gorm:"column:gallery_id"`
	UserID    string `gorm:"column:user_id"`
}

func UserLikeStats(userIDs []string) (map[string]model.UserLikeStats, error) {
	result := map[string]model.UserLikeStats{}
	if len(userIDs) == 0 {
		return result, nil
	}
	db, err := DB()
	if err != nil {
		return result, err
	}
	var givenRows []userLikeCountRow
	if err := db.Model(&model.GalleryLike{}).Select("user_id, COUNT(DISTINCT gallery_id) AS total").Where("user_id IN ?", userIDs).Group("user_id").Scan(&givenRows).Error; err != nil {
		return result, err
	}
	for _, row := range givenRows {
		stats := result[row.UserID]
		stats.Given = int(row.Total)
		result[row.UserID] = stats
	}
	var receivedRows []receivedLikeRow
	err = db.Table("gallery_likes").Select("gallery_images.user_id AS owner_id, gallery_likes.gallery_id, gallery_likes.user_id").Joins("JOIN gallery_images ON gallery_images.id = gallery_likes.gallery_id").Where("gallery_images.user_id IN ?", userIDs).Scan(&receivedRows).Error
	if err != nil {
		return result, err
	}
	seenReceived := map[string]bool{}
	for _, row := range receivedRows {
		key := row.OwnerID + "\x00" + row.GalleryID + "\x00" + row.UserID
		if seenReceived[key] {
			continue
		}
		seenReceived[key] = true
		stats := result[row.OwnerID]
		stats.Received++
		result[row.OwnerID] = stats
	}
	return result, nil
}

// CountUsers 返回用户总数。
func CountUsers() (int64, error) {
	db, err := DB()
	if err != nil {
		return 0, err
	}
	var total int64
	return total, db.Model(&model.User{}).Count(&total).Error
}

// HasAdmin 判断系统中是否存在管理员。
func HasAdmin() (bool, error) {
	db, err := DB()
	if err != nil {
		return false, err
	}
	var total int64
	err = db.Model(&model.User{}).Where("role = ?", model.UserRoleAdmin).Count(&total).Error
	return total > 0, err
}

// GetUserByID 根据 ID 查询用户。
func GetUserByID(id string) (model.User, bool, error) {
	db, err := DB()
	if err != nil {
		return model.User{}, false, err
	}
	return findUser(db, "id = ?", id)
}

// GetUserByUsername 根据用户名查询用户。
func GetUserByUsername(username string) (model.User, bool, error) {
	db, err := DB()
	if err != nil {
		return model.User{}, false, err
	}
	return findUser(db, "username = ?", username)
}

// SaveUser 保存用户信息。
func SaveUser(user model.User) (model.User, error) {
	db, err := DB()
	if err != nil {
		return user, err
	}
	return user, db.Save(&user).Error
}

func ConsumeUserCredits(id string, credits int, now string) (model.User, bool, error) {
	db, err := DB()
	if err != nil {
		return model.User{}, false, err
	}
	if credits <= 0 {
		user, ok, err := GetUserByID(id)
		return user, ok, err
	}
	tx := db.Model(&model.User{}).Where("id = ? AND credits >= ?", id, credits).Updates(map[string]any{
		"credits":    gorm.Expr("credits - ?", credits),
		"updated_at": now,
	})
	if tx.Error != nil {
		return model.User{}, false, tx.Error
	}
	user, ok, err := GetUserByID(id)
	return user, ok && tx.RowsAffected > 0, err
}

func RefundUserCredits(id string, credits int, now string) (model.User, bool, error) {
	db, err := DB()
	if err != nil {
		return model.User{}, false, err
	}
	if credits <= 0 {
		user, ok, err := GetUserByID(id)
		return user, ok, err
	}
	tx := db.Model(&model.User{}).Where("id = ?", id).Updates(map[string]any{
		"credits":    gorm.Expr("credits + ?", credits),
		"updated_at": now,
	})
	if tx.Error != nil {
		return model.User{}, false, tx.Error
	}
	user, ok, err := GetUserByID(id)
	return user, ok && tx.RowsAffected > 0, err
}

func AddUserCreditsWithLog(id string, credits int, now string, log model.CreditLog) (model.User, bool, error) {
	db, err := DB()
	if err != nil {
		return model.User{}, false, err
	}
	if credits < 0 {
		credits = 0
	}
	var user model.User
	err = db.Transaction(func(tx *gorm.DB) error {
		if credits > 0 {
			result := tx.Model(&model.User{}).Where("id = ?", id).Updates(map[string]any{
				"credits":    gorm.Expr("credits + ?", credits),
				"updated_at": now,
			})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
		}
		if err := tx.First(&user, "id = ?", id).Error; err != nil {
			return err
		}
		log.UserID = id
		log.Amount = credits
		log.Balance = user.Credits
		if log.CreatedAt == "" {
			log.CreatedAt = now
		}
		return tx.Save(&log).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.User{}, false, nil
	}
	return user, err == nil, err
}

func CountCreditLogsByType(userID string, logType model.CreditLogType, start string, end string) (int64, error) {
	db, err := DB()
	if err != nil {
		return 0, err
	}
	tx := db.Model(&model.CreditLog{}).Where("user_id = ? AND type = ?", userID, logType)
	if start != "" {
		tx = tx.Where("created_at >= ?", start)
	}
	if end != "" {
		tx = tx.Where("created_at < ?", end)
	}
	var total int64
	return total, tx.Distinct("related_id").Count(&total).Error
}

func CreditLogTypeStats(userID string, logTypes []model.CreditLogType, start string, end string) (map[model.CreditLogType]CreditLogTypeStat, error) {
	result := map[model.CreditLogType]CreditLogTypeStat{}
	if userID == "" || len(logTypes) == 0 {
		return result, nil
	}
	db, err := DB()
	if err != nil {
		return result, err
	}
	tx := db.Model(&model.CreditLog{}).Select("type, COUNT(DISTINCT related_id) AS total, COALESCE(SUM(amount), 0) AS amount").Where("user_id = ? AND type IN ?", userID, logTypes)
	if start != "" {
		tx = tx.Where("created_at >= ?", start)
	}
	if end != "" {
		tx = tx.Where("created_at < ?", end)
	}
	var rows []creditLogTypeStatRow
	if err := tx.Group("type").Scan(&rows).Error; err != nil {
		return result, err
	}
	for _, row := range rows {
		result[row.Type] = CreditLogTypeStat{Type: row.Type, Count: int(row.Count), Amount: row.Amount}
	}
	return result, nil
}

func HasCreditLog(userID string, logType model.CreditLogType, relatedID string) (bool, error) {
	db, err := DB()
	if err != nil {
		return false, err
	}
	var total int64
	err = db.Model(&model.CreditLog{}).Where("user_id = ? AND type = ? AND related_id = ?", userID, logType, relatedID).Count(&total).Error
	return total > 0, err
}

func CheckInUser(id string, date string, credits int, now string, log model.CreditLog) (model.User, bool, error) {
	db, err := DB()
	if err != nil {
		return model.User{}, false, err
	}
	var changed bool
	if err := db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.User{}).Where("id = ? AND status = ? AND (last_check_in_date = '' OR last_check_in_date IS NULL OR last_check_in_date <> ?)", id, model.UserStatusActive, date).Updates(map[string]any{
			"credits":            gorm.Expr("credits + ?", credits),
			"last_check_in_date": date,
			"updated_at":         now,
		})
		if result.Error != nil || result.RowsAffected == 0 {
			return result.Error
		}
		changed = true
		var user model.User
		if err := tx.First(&user, "id = ?", id).Error; err != nil {
			return err
		}
		log.Balance = user.Credits
		return tx.Save(&log).Error
	}); err != nil {
		return model.User{}, false, err
	}
	user, ok, err := GetUserByID(id)
	return user, ok && changed, err
}

// SaveCreditLog 保存算力点变更流水。
func SaveCreditLog(log model.CreditLog) (model.CreditLog, error) {
	db, err := DB()
	if err != nil {
		return log, err
	}
	return log, db.Save(&log).Error
}

func ListCreditLogs(q model.Query) ([]model.AdminCreditLog, int64, error) {
	db, err := DB()
	if err != nil {
		return nil, 0, err
	}
	q.Normalize()
	tx := db.Model(&model.CreditLog{}).Joins("LEFT JOIN users ON users.id = credit_logs.user_id")
	if keyword := strings.TrimSpace(q.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		tx = tx.Where("credit_logs.user_id LIKE ? OR users.username LIKE ? OR users.display_name LIKE ? OR credit_logs.type LIKE ? OR credit_logs.remark LIKE ? OR credit_logs.related_id LIKE ?", like, like, like, like, like, like)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var logs []model.AdminCreditLog
	err = tx.Select("credit_logs.*, users.username, users.display_name").Order("credit_logs.created_at desc").Offset(q.Offset()).Limit(q.PageSize).Scan(&logs).Error
	return logs, total, err
}

func DeleteCreditLog(id string) error {
	db, err := DB()
	if err != nil {
		return err
	}
	return db.Delete(&model.CreditLog{}, "id = ?", id).Error
}

// DeleteUser 删除指定用户。
func DeleteUser(id string) error {
	db, err := DB()
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&model.UserIdentity{}, "user_id = ?", id).Error; err != nil {
			return err
		}
		return tx.Delete(&model.User{}, "id = ?", id).Error
	})
}

// GetUserByLinuxDoID 根据 Linux.do ID 查询用户。
func GetUserByLinuxDoID(id string) (model.User, bool, error) {
	db, err := DB()
	if err != nil {
		return model.User{}, false, err
	}
	return findUser(db, "linux_do_id = ?", id)
}

// findUser 查询单个用户，并将未命中转换为 ok=false。
func findUser(db *gorm.DB, query string, args ...any) (model.User, bool, error) {
	user := model.User{}
	err := db.Where(query, args...).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.User{}, false, nil
	}
	return user, err == nil, err
}
