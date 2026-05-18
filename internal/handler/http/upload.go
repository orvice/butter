package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"go.orx.me/apps/butter/internal/repo/auth"
	"go.orx.me/apps/butter/internal/service"
)

// UploadHandler exposes HTTP endpoints for uploading user/agent avatars and
// other static assets to the configured S3-backed object store.
//
// All endpoints live under /api/uploads so they sit behind the same
// authentication middleware as the rest of the API.
type UploadHandler struct {
	svc *service.UploadService
}

// NewUploadHandler wires the HTTP layer to the upload service. If svc is nil
// (no S3 configured), the handler still registers routes but returns 503.
func NewUploadHandler(svc *service.UploadService) *UploadHandler {
	return &UploadHandler{svc: svc}
}

// Register mounts the upload endpoints on the given engine.
func (h *UploadHandler) Register(r *gin.Engine) {
	g := r.Group("/api/uploads")
	g.POST("/avatar", h.UploadAvatar)
	g.POST("/avatar/:owner_kind/:owner_id", h.UploadAvatarFor)
	g.POST("/static", h.UploadStatic)
}

// UploadAvatar uploads an avatar for the currently authenticated user.
// Expects multipart/form-data with a file field named "file".
func (h *UploadHandler) UploadAvatar(c *gin.Context) {
	if !h.enabled(c) {
		return
	}
	user, ok := auth.UserFromContext(c.Request.Context())
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	h.handleAvatar(c, "user", user.GetId())
}

// UploadAvatarFor uploads an avatar for an arbitrary owner. Admin only when
// the owner is not the caller themself.
//
// Authorization order matters: callers authenticated via the root API token
// (ops/automation) have `auth.IsAdmin` true but no `UserFromContext`. They
// must still be able to set avatars for agents/users, so admin is checked
// first and a session user is only required for the self-upload short-circuit.
func (h *UploadHandler) UploadAvatarFor(c *gin.Context) {
	if !h.enabled(c) {
		return
	}
	ownerKind := c.Param("owner_kind")
	ownerID := c.Param("owner_id")
	if ownerKind == "" || ownerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "owner_kind and owner_id are required"})
		return
	}
	ctx := c.Request.Context()
	if !auth.IsAdmin(ctx) {
		user, ok := auth.UserFromContext(ctx)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
		if !(ownerKind == "user" && user.GetId() == ownerID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin role required"})
			return
		}
	}
	h.handleAvatar(c, ownerKind, ownerID)
}

// UploadStatic uploads an arbitrary static asset. Admin only.
func (h *UploadHandler) UploadStatic(c *gin.Context) {
	if !h.enabled(c) {
		return
	}
	if !auth.IsAdmin(c.Request.Context()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin role required"})
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file field is required"})
		return
	}
	name := strings.TrimSpace(c.PostForm("name"))
	if name == "" {
		name = file.Filename
	}
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	contentType := c.PostForm("content_type")
	if contentType == "" {
		contentType = file.Header.Get("Content-Type")
	}
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot open uploaded file"})
		return
	}
	defer f.Close()

	res, err := h.svc.UploadStatic(c.Request.Context(), name, contentType, f)
	if err != nil {
		writeUploadError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *UploadHandler) handleAvatar(c *gin.Context, ownerKind, ownerID string) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file field is required"})
		return
	}
	contentType := file.Header.Get("Content-Type")
	if ct := c.PostForm("content_type"); ct != "" {
		contentType = ct
	}
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot open uploaded file"})
		return
	}
	defer f.Close()

	res, err := h.svc.UploadAvatar(c.Request.Context(), ownerKind, ownerID, contentType, f)
	if err != nil {
		writeUploadError(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *UploadHandler) enabled(c *gin.Context) bool {
	if h.svc == nil || !h.svc.Enabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "static storage is not configured; set store.s3 and static.s3_bucket in config",
		})
		return false
	}
	return true
}

func writeUploadError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrUploadDisabled):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrUploadTooLarge):
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrUnsupportedContentType):
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
