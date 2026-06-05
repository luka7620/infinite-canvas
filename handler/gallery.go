package handler

import (
	"encoding/json"
	"net/http"

	"github.com/basketikun/infinite-canvas/model"
	"github.com/basketikun/infinite-canvas/service"
)

type publishGalleryImageRequest struct {
	GeneratedImageID string   `json:"generatedImageId"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Tags             []string `json:"tags"`
	ShowPrompt       bool     `json:"showPrompt"`
}

type adminGalleryStatusRequest struct {
	Status model.GalleryStatus `json:"status"`
}

type createGalleryCommentRequest struct {
	Content string `json:"content"`
}

func MyGeneratedImages(w http.ResponseWriter, r *http.Request) {
	user, ok := service.UserFromContext(r.Context())
	if !ok || user.Role == model.UserRoleGuest {
		Fail(w, "未登录或权限不足")
		return
	}
	result, err := service.ListMyGeneratedImageRecords(user.ID, parseQuery(r))
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, result)
}

func PublishGalleryImage(w http.ResponseWriter, r *http.Request) {
	user, ok := service.UserFromContext(r.Context())
	if !ok || user.Role == model.UserRoleGuest {
		Fail(w, "未登录或权限不足")
		return
	}
	var request publishGalleryImageRequest
	_ = json.NewDecoder(r.Body).Decode(&request)
	result, err := service.PublishGalleryImage(user, service.PublishGalleryImageInput{
		GeneratedImageID: request.GeneratedImageID,
		Title:            request.Title,
		Description:      request.Description,
		Tags:             request.Tags,
		ShowPrompt:       request.ShowPrompt,
	})
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, result)
}

func GalleryImages(w http.ResponseWriter, r *http.Request) {
	user, _ := service.UserFromContext(r.Context())
	result, err := service.ListGalleryImages(parseQuery(r), user.ID)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, result)
}

func ToggleGalleryLike(w http.ResponseWriter, r *http.Request, id string) {
	user, ok := service.UserFromContext(r.Context())
	if !ok || user.Role == model.UserRoleGuest {
		Fail(w, "请先登录")
		return
	}
	result, err := service.ToggleGalleryLike(id, user)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, result)
}

func GalleryComments(w http.ResponseWriter, r *http.Request, id string) {
	result, err := service.ListGalleryComments(id, parseQuery(r))
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, result)
}

func CreateGalleryComment(w http.ResponseWriter, r *http.Request, id string) {
	user, ok := service.UserFromContext(r.Context())
	if !ok || user.Role == model.UserRoleGuest {
		Fail(w, "请先登录")
		return
	}
	var request createGalleryCommentRequest
	_ = json.NewDecoder(r.Body).Decode(&request)
	result, err := service.CreateGalleryComment(id, user, service.CreateGalleryCommentInput{Content: request.Content})
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, result)
}

func AdminGalleryImages(w http.ResponseWriter, r *http.Request) {
	result, err := service.ListAdminGalleryImages(parseQuery(r))
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, result)
}

func AdminSaveGalleryImage(w http.ResponseWriter, r *http.Request, id string) {
	var request service.UpdateGalleryImageInput
	_ = json.NewDecoder(r.Body).Decode(&request)
	result, err := service.UpdateAdminGalleryImage(id, request)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, result)
}

func AdminSetGalleryStatus(w http.ResponseWriter, r *http.Request, id string) {
	var request adminGalleryStatusRequest
	_ = json.NewDecoder(r.Body).Decode(&request)
	result, err := service.AdminSetGalleryStatus(id, request.Status)
	if err != nil {
		FailError(w, err)
		return
	}
	OK(w, result)
}

func AdminDeleteGalleryImage(w http.ResponseWriter, r *http.Request, id string) {
	if err := service.DeleteAdminGalleryImage(id); err != nil {
		FailError(w, err)
		return
	}
	OK(w, true)
}
