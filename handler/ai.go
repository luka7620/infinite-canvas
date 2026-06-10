package handler

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/basketikun/infinite-canvas/model"
	"github.com/basketikun/infinite-canvas/service"
)

func AIImagesGenerations(w http.ResponseWriter, r *http.Request) {
	proxyAIRequest(w, r, "/images/generations")
}

func AIImagesEdits(w http.ResponseWriter, r *http.Request) {
	proxyAIRequest(w, r, "/images/edits")
}

func AIChatCompletions(w http.ResponseWriter, r *http.Request) {
	proxyAIRequest(w, r, "/chat/completions")
}

func AIVideos(w http.ResponseWriter, r *http.Request) {
	proxyAIRequest(w, r, "/videos")
}

func AIVideo(w http.ResponseWriter, r *http.Request, id string) {
	proxyAIGetRequest(w, r, "/videos/"+id)
}

func AIVideoContent(w http.ResponseWriter, r *http.Request, id string) {
	proxyAIGetRequest(w, r, "/videos/"+id+"/content")
}

func proxyAIGetRequest(w http.ResponseWriter, r *http.Request, path string) {
	modelName := r.URL.Query().Get("model")
	if strings.TrimSpace(modelName) == "" {
		modelName = "grok-imagine-video"
	}
	channel, err := service.SelectModelChannel(modelName)
	if err != nil {
		log.Printf("AI proxy select channel failed: model=%s err=%v", modelName, err)
		Fail(w, "AI 接口请求失败")
		return
	}
	request, err := http.NewRequest(http.MethodGet, service.BuildModelChannelURL(channel, path), nil)
	if err != nil {
		Fail(w, "AI 接口请求失败")
		return
	}
	request.Header.Set("Authorization", "Bearer "+channel.APIKey)
	copyAIResponse(w, request, nil)
}

func proxyAIRequest(w http.ResponseWriter, r *http.Request, path string) {
	body, contentType, modelName, err := readAIRequest(r)
	if err != nil {
		log.Printf("AI proxy request read failed: %v", err)
		Fail(w, "AI 接口请求失败")
		return
	}
	user, ok := service.UserFromContext(r.Context())
	if !ok {
		Fail(w, "未登录或权限不足")
		return
	}
	credits, err := service.ModelCost(modelName)
	if err != nil {
		log.Printf("AI proxy read model cost failed: model=%s err=%v", modelName, err)
		Fail(w, "AI 接口请求失败")
		return
	}
	credits *= readAIRequestCount(body, contentType)
	channel, err := service.SelectModelChannel(modelName)
	if err != nil {
		log.Printf("AI proxy select channel failed: model=%s err=%v", modelName, err)
		Fail(w, "AI 接口请求失败")
		return
	}
	request, err := http.NewRequest(http.MethodPost, service.BuildModelChannelURL(channel, path), bytes.NewReader(body))
	if err != nil {
		log.Printf("AI proxy build request failed: url=%s err=%v", service.BuildModelChannelURL(channel, path), err)
		Fail(w, "AI 接口请求失败")
		return
	}
	request.Header.Set("Authorization", "Bearer "+channel.APIKey)
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	if err := service.ConsumeUserCredits(user.ID, modelName, credits, path); err != nil {
		FailError(w, err)
		return
	}
	if isImageProxyPath(path) {
		copyAIImageResponse(w, request, func() {
			if err := service.RefundUserCredits(user.ID, modelName, credits, path); err != nil {
				log.Printf("AI proxy refund credits failed: user=%s model=%s credits=%d err=%v", user.ID, modelName, credits, err)
			}
		}, user, body, contentType, modelName, path)
		return
	}
	copyAIResponse(w, request, func() {
		if err := service.RefundUserCredits(user.ID, modelName, credits, path); err != nil {
			log.Printf("AI proxy refund credits failed: user=%s model=%s credits=%d err=%v", user.ID, modelName, credits, err)
		}
	})
}

func copyAIImageResponse(w http.ResponseWriter, request *http.Request, onFailure func(), user model.AuthUser, body []byte, contentType string, modelName string, path string) {
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("AI proxy request failed: url=%s err=%v", request.URL.String(), err)
		if onFailure != nil {
			onFailure()
		}
		Fail(w, "上游连接失败，请稍后重试")
		return
	}
	defer response.Body.Close()

	payload, _ := io.ReadAll(response.Body)
	if response.StatusCode >= http.StatusBadRequest {
		log.Printf("AI upstream error: url=%s status=%d body=%s", request.URL.String(), response.StatusCode, strings.TrimSpace(string(payload[:min(len(payload), 4096)])))
		if onFailure != nil {
			onFailure()
		}
		Fail(w, aiUpstreamErrorMessage(payload, response.StatusCode))
		return
	}

	nextPayload := attachGeneratedImageRecords(payload, user, body, contentType, modelName, path)
	for key, values := range response.Header {
		if strings.EqualFold(key, "Content-Length") {
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)
	_, _ = w.Write(nextPayload)
}

func copyAIResponse(w http.ResponseWriter, request *http.Request, onFailure func()) {
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("AI proxy request failed: url=%s err=%v", request.URL.String(), err)
		if onFailure != nil {
			onFailure()
		}
		Fail(w, "上游连接失败，请稍后重试")
		return
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		payload, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		log.Printf("AI upstream error: url=%s status=%d body=%s", request.URL.String(), response.StatusCode, strings.TrimSpace(string(payload)))
		if onFailure != nil {
			onFailure()
		}
		Fail(w, aiUpstreamErrorMessage(payload, response.StatusCode))
		return
	}

	for key, values := range response.Header {
		if strings.EqualFold(key, "Content-Length") {
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(response.StatusCode)
	_, _ = io.Copy(w, response.Body)
}

func aiUpstreamErrorMessage(payload []byte, statusCode int) string {
	var data map[string]any
	if err := json.Unmarshal(payload, &data); err == nil {
		if msg := jsonString(data["msg"]); msg != "" {
			return msg
		}
		if msg := jsonString(data["message"]); msg != "" {
			return msg
		}
		if msg := jsonString(data["error"]); msg != "" {
			return msg
		}
		if errData, ok := data["error"].(map[string]any); ok {
			if msg := jsonString(errData["message"]); msg != "" {
				return msg
			}
		}
	}
	if statusCode == http.StatusTooManyRequests {
		return "上游限流或额度不足，请稍后重试或检查号池状态"
	}
	if statusCode == 524 {
		return "上游请求超时（524），请稍后重试或减少并发"
	}
	if statusCode >= http.StatusInternalServerError {
		return fmt.Sprintf("上游服务错误（HTTP %d）", statusCode)
	}
	return "AI 接口请求失败"
}

func jsonString(value any) string {
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

func readAIRequest(r *http.Request) ([]byte, string, string, error) {
	contentType := r.Header.Get("Content-Type")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, "", "", err
	}
	modelName := ""
	if strings.HasPrefix(contentType, "multipart/form-data") {
		modelName = readMultipartModel(body, contentType)
	} else {
		var payload struct {
			Model string `json:"model"`
		}
		_ = json.Unmarshal(body, &payload)
		modelName = payload.Model
	}
	if strings.TrimSpace(modelName) == "" {
		return nil, "", "", errMissingModel
	}
	return body, contentType, modelName, nil
}

func readMultipartModel(body []byte, contentType string) string {
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return ""
	}
	reader := multipart.NewReader(bytes.NewReader(body), params["boundary"])
	form, err := reader.ReadForm(32 << 20)
	if err != nil {
		return ""
	}
	defer form.RemoveAll()
	if values := form.Value["model"]; len(values) > 0 {
		return values[0]
	}
	return ""
}

func readAIRequestCount(body []byte, contentType string) int {
	count := 1
	if strings.HasPrefix(contentType, "multipart/form-data") {
		_, params, err := mime.ParseMediaType(contentType)
		if err != nil {
			return count
		}
		form, err := multipart.NewReader(bytes.NewReader(body), params["boundary"]).ReadForm(32 << 20)
		if err != nil {
			return count
		}
		defer form.RemoveAll()
		if values := form.Value["n"]; len(values) > 0 {
			_, _ = fmt.Sscan(values[0], &count)
		}
	} else {
		var payload struct {
			N int `json:"n"`
		}
		_ = json.Unmarshal(body, &payload)
		count = payload.N
	}
	if count < 1 {
		return 1
	}
	return count
}

func attachGeneratedImageRecords(payload []byte, user model.AuthUser, body []byte, contentType string, modelName string, path string) []byte {
	var data map[string]any
	if err := json.Unmarshal(payload, &data); err != nil {
		return payload
	}
	items, ok := data["data"].([]any)
	if !ok {
		return payload
	}
	prompt := readAIRequestValue(body, contentType, "prompt")
	source := firstNonEmpty(readAIRequestValue(body, contentType, "source"), sourceFromAIPath(path))
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		imageURL, mimeType := imageURLFromAIItem(item)
		if imageURL == "" {
			continue
		}
		width, height := imageSizeFromDataURL(imageURL)
		record, err := service.CreateGeneratedImageRecord(service.CreateGeneratedImageRecordInput{
			UserID:   user.ID,
			ImageURL: imageURL,
			Width:    width,
			Height:   height,
			MimeType: mimeType,
			Model:    modelName,
			Prompt:   prompt,
			Source:   source,
		})
		if err != nil {
			log.Printf("save generated image record failed: user=%s model=%s err=%v", user.ID, modelName, err)
			continue
		}
		item["generatedImageId"] = record.ID
	}
	nextPayload, err := json.Marshal(data)
	if err != nil {
		return payload
	}
	return nextPayload
}

func imageURLFromAIItem(item map[string]any) (string, string) {
	if value, ok := item["b64_json"].(string); ok && value != "" {
		return "data:image/png;base64," + value, "image/png"
	}
	if value, ok := item["url"].(string); ok && value != "" {
		return value, ""
	}
	return "", ""
}

func imageSizeFromDataURL(dataURL string) (int, int) {
	const marker = ";base64,"
	index := strings.Index(dataURL, marker)
	if !strings.HasPrefix(dataURL, "data:") || index < 0 {
		return 0, 0
	}
	raw, err := base64.StdEncoding.DecodeString(dataURL[index+len(marker):])
	if err != nil {
		return 0, 0
	}
	config, _, err := image.DecodeConfig(bufio.NewReader(bytes.NewReader(raw)))
	if err != nil {
		return 0, 0
	}
	return config.Width, config.Height
}

func readAIRequestValue(body []byte, contentType string, key string) string {
	if strings.HasPrefix(contentType, "multipart/form-data") {
		return readMultipartValue(body, contentType, key)
	}
	var payload map[string]any
	_ = json.Unmarshal(body, &payload)
	if value, ok := payload[key].(string); ok {
		return value
	}
	return ""
}

func readMultipartValue(body []byte, contentType string, key string) string {
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return ""
	}
	form, err := multipart.NewReader(bytes.NewReader(body), params["boundary"]).ReadForm(32 << 20)
	if err != nil {
		return ""
	}
	defer form.RemoveAll()
	if values := form.Value[key]; len(values) > 0 {
		return values[0]
	}
	return ""
}

func sourceFromAIPath(path string) string {
	if path == "/images/edits" {
		return "canvas-edit"
	}
	return "canvas-node"
}

func isImageProxyPath(path string) bool {
	return path == "/images/generations" || path == "/images/edits"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

var errMissingModel = &aiError{"缺少模型名称"}

type aiError struct {
	message string
}

func (err *aiError) Error() string {
	return err.message
}
