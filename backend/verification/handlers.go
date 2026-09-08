// Package verification provides HTTP handlers for host partner legal requisites verification.
package verification

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"flickey/go-backend/auth"
	"flickey/go-backend/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	unpCompanyRegex    = regexp.MustCompile(`^[0-9]{9}$`)
	unpIndividualRegex = regexp.MustCompile(`^[A-Za-z0-9]{9}$`)
	ibanBelarusRegex   = regexp.MustCompile(`^BY[0-9]{2}[A-Za-z0-9]{4}[0-9]{20}$`)
	bicRegex           = regexp.MustCompile(`^[A-Za-z0-9]{8}([A-Za-z0-9]{3})?$`)
	personalIDRegex    = regexp.MustCompile(`^[0-9]{7}[A-Za-z][0-9]{3}[A-Za-z]{2}[0-9]$`)
)

// Handler holds dependencies for verification endpoints.
type Handler struct {
	DB *gorm.DB
}

// NewHandler creates a new verification Handler.
func NewHandler(database *gorm.DB) *Handler {
	return &Handler{DB: database}
}

// RegisterRoutes registers host verification routes under the given group.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMw gin.HandlerFunc, activeMw gin.HandlerFunc) {
	rg.Use(authMw, activeMw)
	rg.GET("/my", h.GetMyVerification)
	rg.POST("", h.SubmitVerification)
}

// GetMyVerification handles GET /api/v1/verification/my.
func (h *Handler) GetMyVerification(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "MISSING_USER", "message": "User not in context."})
		return
	}

	req, err := db.GetUserVerificationRequest(c.Request.Context(), h.DB, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": err.Error()})
		return
	}

	if req == nil {
		c.JSON(http.StatusOK, VerificationStatusResponse{Status: "none"})
		return
	}

	var reqMap map[string]any
	if req.Requisites != nil {
		_ = json.Unmarshal(req.Requisites, &reqMap)
	}

	var docsList []string
	if req.Documents != nil {
		_ = json.Unmarshal(req.Documents, &docsList)
	}

	c.JSON(http.StatusOK, VerificationStatusResponse{
		ID:              &req.ID,
		Status:          req.Status,
		ProviderType:    req.ProviderType,
		LegalName:       req.LegalName,
		UNP:             req.UNP,
		Requisites:      reqMap,
		Documents:       docsList,
		RejectionReason: req.RejectionReason,
		AdminNote:       req.AdminNote,
		CreatedAt:       &req.CreatedAt,
		UpdatedAt:       &req.UpdatedAt,
	})
}

// SubmitVerification handles POST /api/v1/verification.
func (h *Handler) SubmitVerification(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": "MISSING_USER", "message": "User not in context."})
		return
	}

	var body SubmitVerificationRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "VALIDATION_ERROR", "message": err.Error()})
		return
	}

	// 1. Validate UNP by provider type
	cleanUNP := strings.TrimSpace(body.UNP)
	if body.ProviderType == "individual" || body.ProviderType == "self_employed" {
		if !unpIndividualRegex.MatchString(cleanUNP) {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_UNP", "message": "УНП физлица/самозанятого должен содержать ровно 9 буквенно-цифровых символов."})
			return
		}
	} else {
		if !unpCompanyRegex.MatchString(cleanUNP) {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_UNP", "message": "УНП юридического лица или ИП должен состоять ровно из 9 цифр."})
			return
		}
	}

	// 2. Validate IBAN if provided
	if iban, ok := body.Requisites["iban"].(string); ok && strings.TrimSpace(iban) != "" {
		cleanIBAN := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(iban), " ", ""))
		if !ibanBelarusRegex.MatchString(cleanIBAN) {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_IBAN", "message": "Номер счёта (IBAN) в Беларуси должен состоять из 28 символов и начинаться с BY."})
			return
		}
		body.Requisites["iban"] = cleanIBAN
	}

	// 3. Validate BIC if provided
	if bic, ok := body.Requisites["bic"].(string); ok && strings.TrimSpace(bic) != "" {
		cleanBIC := strings.ToUpper(strings.TrimSpace(bic))
		if !bicRegex.MatchString(cleanBIC) {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_BIC", "message": "БИК банка должен содержать 8 или 11 символов."})
			return
		}
		body.Requisites["bic"] = cleanBIC
	}

	// 4. Validate Personal ID if provided (for individual)
	if personalID, ok := body.Requisites["personal_id"].(string); ok && strings.TrimSpace(personalID) != "" {
		cleanPID := strings.ToUpper(strings.TrimSpace(personalID))
		if !personalIDRegex.MatchString(cleanPID) {
			c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_PERSONAL_ID", "message": "Идентификационный номер документа должен содержать 14 символов (например: 1234567A001PB1)."})
			return
		}
		body.Requisites["personal_id"] = cleanPID
	}

	// Convert requisites and documents to datatypes.JSON
	reqJSON, err := json.Marshal(body.Requisites)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUISITES", "message": "Некорректный формат реквизитов."})
		return
	}

	docsStr := make([]string, len(body.Documents))
	for i, d := range body.Documents {
		docsStr[i] = d.String()
	}
	docsJSON, _ := json.Marshal(docsStr)

	// Save or update existing pending verification request
	txErr := h.DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		existing, err := db.GetUserVerificationRequest(c.Request.Context(), tx, user.ID)
		if err != nil {
			return err
		}

		if existing != nil && (existing.Status == "pending" || existing.Status == "changes_requested") {
			existing.ProviderType = body.ProviderType
			existing.LegalName = strings.TrimSpace(body.LegalName)
			existing.UNP = cleanUNP
			existing.Requisites = datatypes.JSON(reqJSON)
			existing.Documents = datatypes.JSON(docsJSON)
			existing.Status = "pending"
			existing.RejectionReason = ""
			if err := tx.Save(existing).Error; err != nil {
				return err
			}
		} else {
			newReq := &db.VerificationRequest{
				ID:           uuid.New(),
				UserID:       user.ID,
				ProviderType: body.ProviderType,
				LegalName:    strings.TrimSpace(body.LegalName),
				UNP:          cleanUNP,
				Status:       "pending",
				Requisites:   datatypes.JSON(reqJSON),
				Documents:    datatypes.JSON(docsJSON),
			}
			if err := tx.Create(newReq).Error; err != nil {
				return err
			}
		}

		// Update any draft listings of this host that have verification video to awaiting_company_verification
		return tx.Model(&db.Listing{}).
			Where("host_id = ? AND status = 'draft' AND verification_video_id IS NOT NULL", user.ID).
			Update("status", "awaiting_company_verification").Error
	})

	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR", "message": txErr.Error()})
		return
	}

	// Fetch fresh state to return
	h.GetMyVerification(c)
}
