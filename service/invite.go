package service

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/basketikun/infinite-canvas/model"
	"github.com/basketikun/infinite-canvas/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RedeemInviteCodeInput struct {
	Code string `json:"code"`
}

type BatchInviteCodeInput struct {
	Type    model.InviteCodeType `json:"type"`
	Count   int                  `json:"count"`
	Credits int                  `json:"credits"`
	MaxUses int                  `json:"maxUses"`
	Enabled bool                 `json:"enabled"`
	Remark  string               `json:"remark"`
}

func ListInviteCodes(q model.Query) (model.InviteCodeList, error) {
	items, total, err := repository.ListInviteCodes(q)
	if err != nil {
		return model.InviteCodeList{}, err
	}
	return model.InviteCodeList{Items: items, Total: int(total)}, nil
}

func SaveInviteCode(input model.InviteCode) (model.InviteCode, error) {
	item := model.InviteCode{
		ID:      strings.TrimSpace(input.ID),
		Code:    normalizeInviteCode(input.Code),
		Type:    input.Type,
		Credits: input.Credits,
		MaxUses: input.MaxUses,
		Enabled: input.Enabled,
		Remark:  strings.TrimSpace(input.Remark),
	}
	if item.Type == "" {
		item.Type = model.InviteCodeTypeRegister
	}
	if !isValidInviteCodeType(item.Type) {
		return item, safeMessageError{message: "邀请码类型无效"}
	}
	if item.Type == model.InviteCodeTypeRegister {
		item.Credits = 0
	}
	if item.Type == model.InviteCodeTypeCredits && item.Credits <= 0 {
		return item, safeMessageError{message: "兑换额度必须大于 0"}
	}
	if item.Code == "" {
		item.Code = newInviteCode()
	}
	if len(item.Code) < 4 {
		return item, safeMessageError{message: "邀请码至少需要 4 位"}
	}
	if item.MaxUses < 0 {
		item.MaxUses = 0
	}
	if saved, ok, err := repository.GetInviteCodeByCode(item.Code); err != nil {
		return item, err
	} else if ok && saved.ID != item.ID {
		return item, safeMessageError{message: "邀请码已存在"}
	}
	now := now()
	if item.ID == "" {
		item.ID = newID("invite")
		item.CreatedAt = now
	} else if saved, ok, err := repository.GetInviteCodeByID(item.ID); err != nil {
		return item, err
	} else if ok {
		item.CreatedAt = saved.CreatedAt
		item.UsedCount = saved.UsedCount
	}
	item.UpdatedAt = now
	return repository.SaveInviteCode(item)
}

func BatchCreateInviteCodes(input BatchInviteCodeInput) (model.InviteCodeBatchResult, error) {
	if input.Count <= 0 {
		return model.InviteCodeBatchResult{}, safeMessageError{message: "生成数量必须大于 0"}
	}
	if input.Count > 200 {
		return model.InviteCodeBatchResult{}, safeMessageError{message: "单次最多生成 200 个邀请码"}
	}
	items := make([]model.InviteCode, 0, input.Count)
	for range input.Count {
		item, err := SaveInviteCode(model.InviteCode{
			Type:    input.Type,
			Credits: input.Credits,
			MaxUses: input.MaxUses,
			Enabled: input.Enabled,
			Remark:  strings.TrimSpace(input.Remark),
		})
		if err != nil {
			return model.InviteCodeBatchResult{}, err
		}
		items = append(items, item)
	}
	return model.InviteCodeBatchResult{Items: items}, nil
}

func DeleteInviteCode(id string) error {
	return repository.DeleteInviteCode(id)
}

func Register(username string, password string, code string) (model.AuthSession, error) {
	settings, err := repository.GetSettings()
	if err != nil {
		return model.AuthSession{}, err
	}
	normalizedSettings := normalizeSettings(settings)
	allowPublicRegister := normalizedSettings.Public.Auth.AllowRegister == nil || *normalizedSettings.Public.Auth.AllowRegister
	invite, hasInvite, err := validInviteCode(code, model.InviteCodeTypeRegister)
	if err != nil {
		return model.AuthSession{}, err
	}
	if !allowPublicRegister && !hasInvite {
		return model.AuthSession{}, safeMessageError{message: "当前未开放注册，请填写邀请码"}
	}
	username = strings.TrimSpace(username)
	if strings.ContainsAny(username, " \t\r\n") {
		return model.AuthSession{}, safeMessageError{message: "用户名不能包含空格"}
	}
	if username == "" || password == "" {
		return model.AuthSession{}, safeMessageError{message: "用户名和密码不能为空"}
	}
	if _, ok, err := repository.GetUserByUsername(username); err != nil || ok {
		if err != nil {
			return model.AuthSession{}, err
		}
		return model.AuthSession{}, safeMessageError{message: "用户名已存在"}
	}
	hash, err := hashPassword(password)
	if err != nil {
		return model.AuthSession{}, err
	}
	now := now()
	user := model.User{
		ID:        newID("user"),
		Username:  username,
		Password:  hash,
		Role:      model.UserRoleUser,
		AffCode:   newAffCode(),
		Status:    model.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if hasInvite {
		user, invite, err = repository.UseRegisterInviteCode(user, invite, now)
	} else {
		user, err = repository.SaveUser(user)
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.AuthSession{}, safeMessageError{message: "邀请码已用完"}
		}
		return model.AuthSession{}, err
	}
	return newSession(user)
}

func RedeemInviteCode(userID string, input RedeemInviteCodeInput) (model.InviteCodeRedeemResult, error) {
	invite, ok, err := validInviteCode(input.Code, model.InviteCodeTypeCredits)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.InviteCodeRedeemResult{}, safeMessageError{message: "兑换码已用完"}
		}
		return model.InviteCodeRedeemResult{}, err
	}
	if !ok {
		return model.InviteCodeRedeemResult{}, safeMessageError{message: "请输入兑换码"}
	}
	changedAt := now()
	extra, _ := json.Marshal(map[string]string{"code": invite.Code})
	user, invite, changed, err := repository.RedeemInviteCode(userID, invite, changedAt, model.CreditLog{
		ID:        newID("credit"),
		UserID:    userID,
		Type:      model.CreditLogTypeInviteCode,
		Amount:    invite.Credits,
		RelatedID: invite.ID,
		Remark:    "邀请码兑换",
		Extra:     string(extra),
		CreatedAt: changedAt,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.InviteCodeRedeemResult{}, safeMessageError{message: "兑换码已用完"}
		}
		return model.InviteCodeRedeemResult{}, err
	}
	if !changed {
		return model.InviteCodeRedeemResult{}, safeMessageError{message: "该兑换码已使用"}
	}
	authUser := model.PublicUser(user)
	authUser.CheckedInToday = user.LastCheckInDate == checkInDate()
	return model.InviteCodeRedeemResult{User: authUser, InviteCode: invite, Credits: invite.Credits}, nil
}

func validInviteCode(code string, inviteType model.InviteCodeType) (model.InviteCode, bool, error) {
	code = normalizeInviteCode(code)
	if code == "" {
		return model.InviteCode{}, false, nil
	}
	item, ok, err := repository.GetInviteCodeByCode(code)
	if err != nil || !ok {
		if err != nil {
			return item, false, err
		}
		return item, false, safeMessageError{message: "邀请码无效"}
	}
	if item.Type != inviteType {
		return item, false, safeMessageError{message: "邀请码类型不匹配"}
	}
	if !item.Enabled {
		return item, false, safeMessageError{message: "邀请码已停用"}
	}
	if item.MaxUses > 0 && item.UsedCount >= item.MaxUses {
		return item, false, safeMessageError{message: "邀请码已用完"}
	}
	return item, true, nil
}

func requireRegisterInviteCode(code string) (model.InviteCode, error) {
	invite, hasInvite, err := validInviteCode(code, model.InviteCodeTypeRegister)
	if err != nil || hasInvite {
		return invite, err
	}
	return invite, safeMessageError{message: "请先填写注册邀请码"}
}

func normalizeInviteCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func newInviteCode() string {
	raw := strings.ReplaceAll(uuid.NewString(), "-", "")
	return strings.ToUpper(raw[:10])
}

func isValidInviteCodeType(value model.InviteCodeType) bool {
	switch value {
	case model.InviteCodeTypeRegister, model.InviteCodeTypeCredits:
		return true
	default:
		return false
	}
}
