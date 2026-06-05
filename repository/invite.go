package repository

import (
	"errors"
	"strings"

	"github.com/basketikun/infinite-canvas/model"
	"gorm.io/gorm"
)

func ListInviteCodes(q model.Query) ([]model.InviteCode, int64, error) {
	db, err := DB()
	if err != nil {
		return nil, 0, err
	}
	q.Normalize()
	tx := db.Model(&model.InviteCode{})
	if keyword := strings.TrimSpace(q.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		tx = tx.Where("code LIKE ? OR remark LIKE ?", like, like)
	}
	if isInviteCodeType(q.Type) {
		tx = tx.Where("type = ?", q.Type)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.InviteCode
	err = tx.Order("created_at desc").Offset(q.Offset()).Limit(q.PageSize).Find(&items).Error
	return items, total, err
}

func GetInviteCodeByID(id string) (model.InviteCode, bool, error) {
	db, err := DB()
	if err != nil {
		return model.InviteCode{}, false, err
	}
	return findInviteCode(db, "id = ?", id)
}

func GetInviteCodeByCode(code string) (model.InviteCode, bool, error) {
	db, err := DB()
	if err != nil {
		return model.InviteCode{}, false, err
	}
	return findInviteCode(db, "code = ?", code)
}

func SaveInviteCode(item model.InviteCode) (model.InviteCode, error) {
	db, err := DB()
	if err != nil {
		return item, err
	}
	return item, db.Save(&item).Error
}

func DeleteInviteCode(id string) error {
	db, err := DB()
	if err != nil {
		return err
	}
	return db.Delete(&model.InviteCode{}, "id = ?", id).Error
}

func CreateUserWithCreditLog(user model.User, log *model.CreditLog) (model.User, error) {
	db, err := DB()
	if err != nil {
		return user, err
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&user).Error; err != nil {
			return err
		}
		if log != nil {
			return tx.Save(log).Error
		}
		return nil
	}); err != nil {
		return user, err
	}
	return user, nil
}

func UseRegisterInviteCode(user model.User, invite model.InviteCode, now string, log *model.CreditLog) (model.User, model.InviteCode, error) {
	db, err := DB()
	if err != nil {
		return user, invite, err
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&user).Error; err != nil {
			return err
		}
		result := tx.Model(&model.InviteCode{}).Where("id = ? AND enabled = ? AND type = ? AND (max_uses <= 0 OR used_count < max_uses)", invite.ID, true, model.InviteCodeTypeRegister).Updates(map[string]any{
			"used_count": gorm.Expr("used_count + ?", 1),
			"updated_at": now,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		use := model.InviteCodeUse{
			ID:           user.ID + "-" + invite.ID,
			InviteCodeID: invite.ID,
			Code:         invite.Code,
			UserID:       user.ID,
			Type:         invite.Type,
			Credits:      invite.Credits,
			CreatedAt:    now,
		}
		if err := tx.Save(&use).Error; err != nil {
			return err
		}
		if log != nil {
			if err := tx.Save(log).Error; err != nil {
				return err
			}
		}
		return tx.First(&invite, "id = ?", invite.ID).Error
	}); err != nil {
		return user, invite, err
	}
	return user, invite, nil
}

func CreateLinuxDoUserWithInviteCode(user model.User, invite model.InviteCode, now string, log *model.CreditLog) (model.User, model.InviteCode, error) {
	db, err := DB()
	if err != nil {
		return user, invite, err
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&user).Error; err != nil {
			return err
		}
		identity := model.UserIdentity{
			ID:         string(model.UserIdentityProviderLinuxDo) + "-" + user.LinuxDoID,
			UserID:     user.ID,
			Provider:   model.UserIdentityProviderLinuxDo,
			ExternalID: user.LinuxDoID,
			CreatedAt:  now,
		}
		if err := deleteOrphanUserIdentity(tx, identity.ID); err != nil {
			return err
		}
		if err := tx.Create(&identity).Error; err != nil {
			return err
		}
		result := tx.Model(&model.InviteCode{}).Where("id = ? AND enabled = ? AND type = ? AND (max_uses <= 0 OR used_count < max_uses)", invite.ID, true, model.InviteCodeTypeRegister).Updates(map[string]any{
			"used_count": gorm.Expr("used_count + ?", 1),
			"updated_at": now,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		use := model.InviteCodeUse{
			ID:           user.ID + "-" + invite.ID,
			InviteCodeID: invite.ID,
			Code:         invite.Code,
			UserID:       user.ID,
			Type:         invite.Type,
			Credits:      invite.Credits,
			CreatedAt:    now,
		}
		if err := tx.Save(&use).Error; err != nil {
			return err
		}
		if log != nil {
			if err := tx.Save(log).Error; err != nil {
				return err
			}
		}
		return tx.First(&invite, "id = ?", invite.ID).Error
	}); err != nil {
		return user, invite, err
	}
	return user, invite, nil
}

func deleteOrphanUserIdentity(tx *gorm.DB, id string) error {
	var identity model.UserIdentity
	if err := tx.First(&identity, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	var total int64
	if err := tx.Model(&model.User{}).Where("id = ?", identity.UserID).Count(&total).Error; err != nil {
		return err
	}
	if total > 0 {
		return nil
	}
	return tx.Delete(&model.UserIdentity{}, "id = ?", id).Error
}

func RedeemInviteCode(userID string, invite model.InviteCode, now string, log model.CreditLog) (model.User, model.InviteCode, bool, error) {
	db, err := DB()
	if err != nil {
		return model.User{}, invite, false, err
	}
	used := false
	if err := db.Transaction(func(tx *gorm.DB) error {
		var existing int64
		if err := tx.Model(&model.InviteCodeUse{}).Where("invite_code_id = ? AND user_id = ?", invite.ID, userID).Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return nil
		}
		used = true
		result := tx.Model(&model.InviteCode{}).Where("id = ? AND enabled = ? AND type = ? AND (max_uses <= 0 OR used_count < max_uses)", invite.ID, true, model.InviteCodeTypeCredits).Updates(map[string]any{
			"used_count": gorm.Expr("used_count + ?", 1),
			"updated_at": now,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		if err := tx.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]any{
			"credits":    gorm.Expr("credits + ?", invite.Credits),
			"updated_at": now,
		}).Error; err != nil {
			return err
		}
		var user model.User
		if err := tx.First(&user, "id = ?", userID).Error; err != nil {
			return err
		}
		use := model.InviteCodeUse{
			ID:           log.ID + "-use",
			InviteCodeID: invite.ID,
			Code:         invite.Code,
			UserID:       userID,
			Type:         invite.Type,
			Credits:      invite.Credits,
			CreatedAt:    now,
		}
		if err := tx.Save(&use).Error; err != nil {
			return err
		}
		log.Balance = user.Credits
		if err := tx.Save(&log).Error; err != nil {
			return err
		}
		return tx.First(&invite, "id = ?", invite.ID).Error
	}); err != nil {
		return model.User{}, invite, false, err
	}
	user, ok, err := GetUserByID(userID)
	return user, invite, ok && used, err
}

func findInviteCode(db *gorm.DB, query string, args ...any) (model.InviteCode, bool, error) {
	item := model.InviteCode{}
	err := db.Where(query, args...).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.InviteCode{}, false, nil
	}
	return item, err == nil, err
}

func isInviteCodeType(value string) bool {
	switch model.InviteCodeType(value) {
	case model.InviteCodeTypeRegister, model.InviteCodeTypeCredits:
		return true
	default:
		return false
	}
}
