package router

import (
	"net/http"

	"github.com/basketikun/infinite-canvas/handler"
	"github.com/basketikun/infinite-canvas/middleware"
	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	router := gin.Default()
	router.RedirectTrailingSlash = false
	_ = router.SetTrustedProxies(nil)
	api := router.Group("/api")
	api.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	api.POST("/auth/register", gin.WrapF(handler.Register))
	api.POST("/auth/login", gin.WrapF(handler.Login))
	api.GET("/auth/linux-do/authorize", gin.WrapF(handler.LinuxDoAuthorize))
	api.GET("/auth/linux-do/callback", gin.WrapF(handler.LinuxDoCallback))
	api.GET("/auth/me", middleware.OptionalAuth, gin.WrapF(handler.CurrentUser))
	api.POST("/auth/check-in", middleware.UserAuth, gin.WrapF(handler.CheckIn))
	api.POST("/auth/invite-codes/redeem", middleware.UserAuth, gin.WrapF(handler.RedeemInviteCode))
	api.GET("/settings", gin.WrapF(handler.Settings))
	v1 := api.Group("/v1", middleware.UserAuth)
	v1.POST("/images/generations", gin.WrapF(handler.AIImagesGenerations))
	v1.POST("/images/edits", gin.WrapF(handler.AIImagesEdits))
	v1.POST("/chat/completions", gin.WrapF(handler.AIChatCompletions))
	v1.POST("/videos", gin.WrapF(handler.AIVideos))
	v1.GET("/videos/:id", func(c *gin.Context) {
		handler.AIVideo(c.Writer, c.Request, c.Param("id"))
	})
	v1.GET("/videos/:id/content", func(c *gin.Context) {
		handler.AIVideoContent(c.Writer, c.Request, c.Param("id"))
	})
	api.GET("/prompts", middleware.OptionalAuth, gin.WrapF(handler.Prompts))
	api.GET("/assets", middleware.OptionalAuth, gin.WrapF(handler.Assets))
	api.GET("/gallery", middleware.OptionalAuth, gin.WrapF(handler.GalleryImages))
	api.GET("/gallery/:id/comments", func(c *gin.Context) {
		handler.GalleryComments(c.Writer, c.Request, c.Param("id"))
	})
	api.GET("/generated-images", middleware.UserAuth, gin.WrapF(handler.MyGeneratedImages))
	api.POST("/gallery", middleware.UserAuth, gin.WrapF(handler.PublishGalleryImage))
	api.POST("/gallery/:id/like", middleware.UserAuth, func(c *gin.Context) {
		handler.ToggleGalleryLike(c.Writer, c.Request, c.Param("id"))
	})
	api.POST("/gallery/:id/comments", middleware.UserAuth, func(c *gin.Context) {
		handler.CreateGalleryComment(c.Writer, c.Request, c.Param("id"))
	})
	api.POST("/admin/login", gin.WrapF(handler.AdminLogin))

	admin := api.Group("/admin", middleware.AdminAuth)
	admin.GET("/stats", gin.WrapF(handler.AdminStats))
	admin.GET("/users", gin.WrapF(handler.AdminUsers))
	admin.POST("/users", gin.WrapF(handler.AdminSaveUser))
	admin.POST("/users/:id/credits", func(c *gin.Context) {
		handler.AdminAdjustUserCredits(c.Writer, c.Request, c.Param("id"))
	})
	admin.DELETE("/users/:id", func(c *gin.Context) {
		handler.AdminDeleteUser(c.Writer, c.Request, c.Param("id"))
	})
	admin.GET("/credit-logs", gin.WrapF(handler.AdminCreditLogs))
	admin.POST("/credit-logs", gin.WrapF(handler.AdminSaveCreditLog))
	admin.DELETE("/credit-logs/:id", func(c *gin.Context) {
		handler.AdminDeleteCreditLog(c.Writer, c.Request, c.Param("id"))
	})
	admin.GET("/invite-codes", gin.WrapF(handler.AdminInviteCodes))
	admin.POST("/invite-codes", gin.WrapF(handler.AdminSaveInviteCode))
	admin.POST("/invite-codes/batch", gin.WrapF(handler.AdminBatchInviteCodes))
	admin.DELETE("/invite-codes/:id", func(c *gin.Context) {
		handler.AdminDeleteInviteCode(c.Writer, c.Request, c.Param("id"))
	})
	admin.GET("/settings", gin.WrapF(handler.AdminSettings))
	admin.POST("/settings", gin.WrapF(handler.AdminSaveSettings))
	admin.POST("/settings/channel-models", gin.WrapF(handler.AdminChannelModels))
	admin.POST("/settings/channel-test", gin.WrapF(handler.AdminTestChannelModel))
	admin.GET("/prompt-categories", gin.WrapF(handler.AdminPromptCategories))
	admin.POST("/prompt-categories/sync", gin.WrapF(handler.AdminSyncPromptCategories))
	admin.GET("/prompts", gin.WrapF(handler.AdminPrompts))
	admin.POST("/prompts", gin.WrapF(handler.AdminSavePrompt))
	admin.POST("/prompts/batch-delete", gin.WrapF(handler.AdminDeletePrompts))
	admin.DELETE("/prompts/:id", func(c *gin.Context) {
		handler.AdminDeletePrompt(c.Writer, c.Request, c.Param("id"))
	})
	admin.GET("/assets", gin.WrapF(handler.AdminAssets))
	admin.POST("/assets", gin.WrapF(handler.AdminSaveAsset))
	admin.DELETE("/assets/:id", func(c *gin.Context) {
		handler.AdminDeleteAsset(c.Writer, c.Request, c.Param("id"))
	})
	admin.GET("/gallery", gin.WrapF(handler.AdminGalleryImages))
	admin.POST("/gallery/:id", func(c *gin.Context) {
		handler.AdminSaveGalleryImage(c.Writer, c.Request, c.Param("id"))
	})
	admin.POST("/gallery/:id/status", func(c *gin.Context) {
		handler.AdminSetGalleryStatus(c.Writer, c.Request, c.Param("id"))
	})
	admin.DELETE("/gallery/:id", func(c *gin.Context) {
		handler.AdminDeleteGalleryImage(c.Writer, c.Request, c.Param("id"))
	})

	router.NoRoute(middleware.NotFoundJSON)

	return router
}
