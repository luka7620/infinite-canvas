package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/basketikun/infinite-canvas/model"
	"github.com/basketikun/infinite-canvas/service"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

type saveUserRequest struct {
	ID          string           `json:"id"`
	Username    string           `json:"username"`
	Password    string           `json:"password"`
	Email       string           `json:"email"`
	DisplayName string           `json:"displayName"`
	AvatarURL   string           `json:"avatarUrl"`
	Role        model.UserRole   `json:"role"`
	Status      model.UserStatus `json:"status"`
}

type adjustUserCreditsRequest struct {
	Credits int `json:"credits"`
}

type redeemInviteCodeRequest struct {
	Code string `json:"code"`
}

type batchInviteCodeRequest struct {
	Type    model.InviteCodeType `json:"type"`
	Count   int                  `json:"count"`
	Credits int                  `json:"credits"`
	MaxUses int                  `json:"maxUses"`
	Enabled bool                 `json:"enabled"`
	Remark  string               `json:"remark"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	var request registerRequest
	_ = json.NewDecoder(r.Body).Decode(&request)
	session, err := service.Register(request.Username, request.Password, request.Code)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, session)
}

func Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	_ = json.NewDecoder(r.Body).Decode(&request)
	session, err := service.Login(request.Username, request.Password)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, session)
}

func LinuxDoAuthorize(w http.ResponseWriter, r *http.Request) {
	redirect := r.URL.Query().Get("redirect")
	authURL, err := service.LinuxDoAuthorizeURL(r, redirect, r.URL.Query().Get("code"))
	if err != nil {
		http.Redirect(w, r, loginRedirect(r, redirect, "", err.Error()), http.StatusFound)
		return
	}
	http.Redirect(w, r, authURL, http.StatusFound)
}

func LinuxDoCallback(w http.ResponseWriter, r *http.Request) {
	session, redirect, err := service.LoginWithLinuxDo(r, r.URL.Query().Get("code"), r.URL.Query().Get("state"))
	if err != nil {
		http.Redirect(w, r, loginRedirect(r, redirect, "", err.Error()), http.StatusFound)
		return
	}
	http.Redirect(w, r, loginRedirect(r, redirect, session.Token, ""), http.StatusFound)
}

func AdminLogin(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	_ = json.NewDecoder(r.Body).Decode(&request)
	session, err := service.Login(request.Username, request.Password)
	if err != nil {
		FailError(w, err)
		return
	}
	if session.User.Role != model.UserRoleAdmin {
		Fail(w, "需要管理员权限")
		return
	}
	OK(w, session)
}

func CurrentUser(w http.ResponseWriter, r *http.Request) {
	if user, ok := service.UserFromContext(r.Context()); ok {
		OK(w, user)
		return
	}
	OK(w, service.GuestUser())
}

func CheckIn(w http.ResponseWriter, r *http.Request) {
	user, ok := service.UserFromContext(r.Context())
	if !ok || user.Role == model.UserRoleGuest {
		Fail(w, "未登录或权限不足")
		return
	}
	result, err := service.CheckIn(user.ID)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, result)
}

func RedeemInviteCode(w http.ResponseWriter, r *http.Request) {
	user, ok := service.UserFromContext(r.Context())
	if !ok || user.Role == model.UserRoleGuest {
		Fail(w, "未登录或权限不足")
		return
	}
	var request redeemInviteCodeRequest
	_ = json.NewDecoder(r.Body).Decode(&request)
	result, err := service.RedeemInviteCode(user.ID, service.RedeemInviteCodeInput{Code: request.Code})
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, result)
}

func SaveProfile(w http.ResponseWriter, r *http.Request) {
	current, ok := service.UserFromContext(r.Context())
	if !ok || current.Role == model.UserRoleGuest {
		Fail(w, "未登录或权限不足")
		return
	}
	var request saveUserRequest
	_ = json.NewDecoder(r.Body).Decode(&request)
	user, err := service.SaveProfile(current.ID, service.ProfileInput{
		Username:    request.Username,
		DisplayName: request.DisplayName,
		AvatarURL:   request.AvatarURL,
	})
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, user)
}

func AdminUsers(w http.ResponseWriter, r *http.Request) {
	users, err := service.ListUsers(parseQuery(r))
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, users)
}

func AdminSaveUser(w http.ResponseWriter, r *http.Request) {
	var request saveUserRequest
	_ = json.NewDecoder(r.Body).Decode(&request)
	user, err := service.SaveUser(model.User{
		ID:          request.ID,
		Username:    request.Username,
		Email:       request.Email,
		DisplayName: request.DisplayName,
		AvatarURL:   request.AvatarURL,
		Role:        request.Role,
		Status:      request.Status,
	}, request.Password)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, user)
}

func AdminAdjustUserCredits(w http.ResponseWriter, r *http.Request, id string) {
	var request adjustUserCreditsRequest
	_ = json.NewDecoder(r.Body).Decode(&request)
	user, err := service.AdjustUserCredits(id, request.Credits)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, user)
}

func AdminCreditLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := service.ListCreditLogs(parseQuery(r))
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, logs)
}

func AdminSaveCreditLog(w http.ResponseWriter, r *http.Request) {
	var log model.CreditLog
	_ = json.NewDecoder(r.Body).Decode(&log)
	result, err := service.SaveCreditLog(log)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, result)
}

func AdminDeleteCreditLog(w http.ResponseWriter, r *http.Request, id string) {
	if err := service.DeleteCreditLog(id); err != nil {
		FailError(w, err)
		return
	}
	OK(w, true)
}

func AdminInviteCodes(w http.ResponseWriter, r *http.Request) {
	items, err := service.ListInviteCodes(parseQuery(r))
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, items)
}

func AdminSaveInviteCode(w http.ResponseWriter, r *http.Request) {
	var item model.InviteCode
	_ = json.NewDecoder(r.Body).Decode(&item)
	result, err := service.SaveInviteCode(item)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, result)
}

func AdminBatchInviteCodes(w http.ResponseWriter, r *http.Request) {
	var request batchInviteCodeRequest
	_ = json.NewDecoder(r.Body).Decode(&request)
	result, err := service.BatchCreateInviteCodes(service.BatchInviteCodeInput(request))
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, result)
}

func AdminDeleteInviteCode(w http.ResponseWriter, r *http.Request, id string) {
	if err := service.DeleteInviteCode(id); err != nil {
		FailError(w, err)
		return
	}
	OK(w, true)
}

func loginRedirect(r *http.Request, redirect string, token string, message string) string {
	values := url.Values{}
	if strings.TrimSpace(token) != "" {
		values.Set("token", token)
	}
	if strings.TrimSpace(message) != "" {
		values.Set("error", message)
	}
	if strings.TrimSpace(redirect) != "" {
		values.Set("redirect", redirect)
	}
	return service.RequestOrigin(r) + "/login?" + values.Encode()
}

func AdminDeleteUser(w http.ResponseWriter, r *http.Request, id string) {
	if err := service.DeleteUser(id); err != nil {
		FailError(w, err)
		return
	}
	OK(w, true)
}
