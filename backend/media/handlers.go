// Package media provides HTTP handlers for media upload operations.
package media

import (
	"io"
	"net/http"

	"flickey/go-backend/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler holds dependencies for media HTTP handlers.
type Handler struct {
	Service *MediaService
}

// NewHandler creates a media Handler.
func NewHandler(svc *MediaService) *Handler {
	return &Handler{Service: svc}
}

// RegisterRoutes registers media routes under the provided RouterGroup.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMw gin.HandlerFunc, activeMw gin.HandlerFunc) {
	active := rg.Group("")
	active.Use(authMw, activeMw)
	active.POST("/presign", h.Presign)
	active.POST("/:media_id/complete", h.Complete)

	// Dev upload & download routes — no auth required.
	rg.PUT("/dev-upload/*file_key", h.DevUpload)
	rg.GET("/dev-upload/*file_key", h.DevGet)
}

// Presign handles POST /media/presign.
//
//	@Summary		Request presigned upload URL
//	@Description	Validates content type and file size, creates a media record with 'pending' status, and returns a presigned upload URL.
//	@Tags			Media
//	@Accept			json
//	@Produce		json
//	@Param			body	body		PresignRequest		true	"Media upload request parameters (content_type and file_size_bytes)"
//	@Success		200		{object}	PresignResponse		"Presigned upload URL and media metadata"
//	@Failure		400		{object}	auth.ErrorResponse	"VALIDATION_ERROR / INVALID_CONTENT_TYPE / INVALID_FILE_SIZE"
//	@Failure		401		{object}	auth.ErrorResponse	"MISSING_TOKEN / INVALID_TOKEN"
//	@Failure		403		{object}	auth.ErrorResponse	"PROFILE_INCOMPLETE"
//	@Failure		500		{object}	auth.ErrorResponse	"INTERNAL_ERROR"
//	@Security		BearerAuth
//	@Router			/media/presign [post]
func (h *Handler) Presign(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "MISSING_USER", "message": "User not in context."})
		return
	}

	var req PresignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": err.Error()})
		return
	}

	resp, err := h.Service.CreatePresignedUpload(c.Request.Context(), user.ID, req.ContentType, req.FileSizeBytes)
	if err != nil {
		handleMediaError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Complete handles POST /media/:media_id/complete.
//
//	@Summary		Complete media upload
//	@Description	Verifies uploaded object existence, size, and binary magic bytes in storage. Marks media status as 'uploaded'.
//	@Tags			Media
//	@Produce		json
//	@Param			media_id	path		string				true	"Media UUID"	format(uuid)
//	@Success		200			{object}	CompleteResponse	"Upload confirmation and file key"
//	@Failure		400			{object}	auth.ErrorResponse	"INVALID_MEDIA_ID / INVALID_MEDIA_STATUS"
//	@Failure		401			{object}	auth.ErrorResponse	"MISSING_TOKEN / INVALID_TOKEN"
//	@Failure		403			{object}	auth.ErrorResponse	"PROFILE_INCOMPLETE / MEDIA_NOT_OWNED"
//	@Failure		422			{object}	auth.ErrorResponse	"OBJECT_NOT_FOUND / INVALID_FILE_SIZE / INVALID_FILE_SIGNATURE"
//	@Failure		500			{object}	auth.ErrorResponse	"INTERNAL_ERROR"
//	@Security		BearerAuth
//	@Router			/media/{media_id}/complete [post]
func (h *Handler) Complete(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "MISSING_USER", "message": "User not in context."})
		return
	}

	mediaIDStr := c.Param("media_id")
	mediaID, err := uuid.Parse(mediaIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_MEDIA_ID", "message": "Invalid media ID."})
		return
	}

	resp, err := h.Service.CompleteUpload(c.Request.Context(), user.ID, mediaID)
	if err != nil {
		handleMediaError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DevUpload handles PUT /media/dev-upload/*file_key.
// Stores file directly in local dev storage. No auth required (dev only).
//
//	@Summary		Dev direct file upload
//	@Description	Uploads file content directly to local development storage without S3. **Dev only.**
//	@Tags			Media
//	@Accept			octet-stream
//	@Produce		json
//	@Param			file_key	path	string	true	"Storage file path/key"
//	@Success		200			"OK"
//	@Failure		400			{object}	auth.ErrorResponse	"READ_ERROR"
//	@Failure		500			{object}	auth.ErrorResponse	"STORAGE_ERROR"
//	@Router			/media/dev-upload/{file_key} [put]
func (h *Handler) DevUpload(c *gin.Context) {
	fileKey := c.Param("file_key")
	if len(fileKey) > 0 && fileKey[0] == '/' {
		fileKey = fileKey[1:]
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "READ_ERROR", "message": "Failed to read request body."})
		return
	}

	contentType := c.GetHeader("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	if err := h.Service.Storage.PutObject(c.Request.Context(), fileKey, body, contentType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "STORAGE_ERROR", "message": "Failed to store file."})
		return
	}

	c.Status(http.StatusOK)
}

// DevGet handles GET /media/dev-upload/*file_key.
// Serves stored file directly to the browser for dev mode.
//
//	@Summary		Dev get uploaded file
//	@Description	Retrieves raw media bytes directly from local dev storage. **Dev only.**
//	@Tags			Media
//	@Produce		image/jpeg
//	@Produce		image/png
//	@Produce		image/webp
//	@Param			file_key	path	string	true	"Storage file path/key"
//	@Success		200			{file}	binary
//	@Failure		404			{object}	auth.ErrorResponse	"FILE_NOT_FOUND"
//	@Router			/media/dev-upload/{file_key} [get]
func (h *Handler) DevGet(c *gin.Context) {
	fileKey := c.Param("file_key")
	if len(fileKey) > 0 && fileKey[0] == '/' {
		fileKey = fileKey[1:]
	}

	data, err := h.Service.Storage.GetObjectBytes(c.Request.Context(), fileKey, 50*1024*1024)
	if err != nil || len(data) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": "FILE_NOT_FOUND", "message": "File not found in storage."})
		return
	}

	contentType := http.DetectContentType(data)
	c.Data(http.StatusOK, contentType, data)
}

func handleMediaError(c *gin.Context, err error) {
	switch e := err.(type) {
	case *ValidationError:
		c.JSON(http.StatusBadRequest, gin.H{"code": e.Code, "message": e.Message})
	case *NotOwnedError:
		c.JSON(http.StatusForbidden, gin.H{"code": e.Code, "message": e.Message})
	case *StorageError:
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": e.Code, "message": e.Message})
	case *InvalidMediaError:
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": e.Code, "message": e.Message})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": "An internal error occurred."})
	}
}
