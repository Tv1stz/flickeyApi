// Package admin provides administrative moderation, compliance, and user enforcement services.
package admin

import (
	"context"
	"fmt"
	"time"

	"flickey/go-backend/db"
	"flickey/go-backend/notifications"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminService struct {
	DB            *gorm.DB
	Notifications *notifications.NotificationService
}

func NewAdminService(database *gorm.DB, notifSvc *notifications.NotificationService) *AdminService {
	return &AdminService{DB: database, Notifications: notifSvc}
}

// ─────────────────────────────────────────────────────────────────────────────
// Audit Log Helper
// ─────────────────────────────────────────────────────────────────────────────

func (s *AdminService) CreateAuditLog(ctx context.Context, adminID uuid.UUID, adminName, action, targetType, targetID, reason, note, oldStatus, newStatus string) error {
	log := &db.AuditLog{
		ID:         uuid.New(),
		AdminID:    adminID,
		AdminName:  adminName,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Reason:     reason,
		Note:       note,
		OldStatus:  oldStatus,
		NewStatus:  newStatus,
		CreatedAt:  time.Now(),
	}
	return s.DB.WithContext(ctx).Create(log).Error
}

// ─────────────────────────────────────────────────────────────────────────────
// Overview & Metrics
// ─────────────────────────────────────────────────────────────────────────────

type OverviewMetrics struct {
	TotalListings      int64 `json:"total_listings"`
	PendingListings    int64 `json:"pending_listings"`
	ApprovedListings   int64 `json:"approved_listings"`
	RejectedListings   int64 `json:"rejected_listings"`
	TotalUsers         int64 `json:"total_users"`
	ActiveUsers        int64 `json:"active_users"`
	BlockedUsers       int64 `json:"blocked_users"`
	PendingVerifications int64 `json:"pending_verifications"`
	OpenReports        int64 `json:"open_reports"`
}

func (s *AdminService) GetOverview(ctx context.Context) (*OverviewMetrics, error) {
	var metrics OverviewMetrics

	s.DB.WithContext(ctx).Model(&db.Listing{}).Count(&metrics.TotalListings)
	s.DB.WithContext(ctx).Model(&db.Listing{}).Where("status IN ?", []string{"pending_review", "awaiting_company_verification"}).Count(&metrics.PendingListings)
	s.DB.WithContext(ctx).Model(&db.Listing{}).Where("status = ?", "published").Count(&metrics.ApprovedListings)
	s.DB.WithContext(ctx).Model(&db.Listing{}).Where("status = ?", "rejected").Count(&metrics.RejectedListings)

	s.DB.WithContext(ctx).Model(&db.User{}).Count(&metrics.TotalUsers)
	s.DB.WithContext(ctx).Model(&db.User{}).Where("status = ?", "active").Count(&metrics.ActiveUsers)
	s.DB.WithContext(ctx).Model(&db.User{}).Where("status IN ?", []string{"suspended", "banned"}).Count(&metrics.BlockedUsers)

	s.DB.WithContext(ctx).Model(&db.VerificationRequest{}).Where("status = ?", "pending").Count(&metrics.PendingVerifications)
	s.DB.WithContext(ctx).Model(&db.Report{}).Where("status = ?", "open").Count(&metrics.OpenReports)

	return &metrics, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Listing Moderation
// ─────────────────────────────────────────────────────────────────────────────

func (s *AdminService) GetListings(ctx context.Context, status string) ([]db.Listing, error) {
	var listings []db.Listing
	query := s.DB.WithContext(ctx).Preload("Host").Preload("ListingAmenities").Preload("Media").Order("created_at DESC")
	if status == "pending_review" || status == "pending" {
		query = query.Where("status IN ?", []string{"pending_review", "awaiting_company_verification"})
	} else if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Find(&listings).Error; err != nil {
		return nil, err
	}
	return listings, nil
}

func (s *AdminService) GetListingByID(ctx context.Context, listingID uuid.UUID) (*db.Listing, error) {
	var listing db.Listing
	if err := s.DB.WithContext(ctx).Preload("Host").Preload("ListingAmenities").Preload("Media").Where("id = ?", listingID).First(&listing).Error; err != nil {
		return nil, fmt.Errorf("listing not found: %w", err)
	}
	return &listing, nil
}

func (s *AdminService) ModerateListing(ctx context.Context, admin *db.User, listingID uuid.UUID, action, reason, note string) (*db.Listing, error) {
	var listing db.Listing
	if err := s.DB.WithContext(ctx).Where("id = ?", listingID).First(&listing).Error; err != nil {
		return nil, fmt.Errorf("listing not found: %w", err)
	}

	oldStatus := listing.Status
	newStatus := oldStatus

	switch action {
	case "approve":
		newStatus = "published"
	case "reject":
		newStatus = "rejected"
	case "request_changes":
		newStatus = "changes_requested"
	case "suspend":
		newStatus = "suspended"
	default:
		return nil, fmt.Errorf("invalid moderation action: %s", action)
	}

	listing.Status = newStatus
	if err := s.DB.WithContext(ctx).Save(&listing).Error; err != nil {
		return nil, err
	}

	adminName := "Admin"
	if admin.FirstName != nil {
		adminName = *admin.FirstName
	}

	_ = s.CreateAuditLog(ctx, admin.ID, adminName, "listing_"+action, "listing", listingID.String(), reason, note, oldStatus, newStatus)

	if s.Notifications != nil {
		switch action {
		case "approve":
			_, _ = s.Notifications.CreateNotification(
				ctx,
				listing.HostID,
				notifications.TypeListingApproved,
				"Объявление опубликовано",
				fmt.Sprintf("Ваше объявление «%s» успешно прошло проверку и опубликовано.", listing.Name),
				map[string]any{"listing_id": listing.ID},
			)
		case "reject":
			msg := fmt.Sprintf("Ваше объявление «%s» отклонено модератором.", listing.Name)
			if reason != "" {
				msg += " Причина: " + reason
			}
			_, _ = s.Notifications.CreateNotification(
				ctx,
				listing.HostID,
				notifications.TypeListingRejected,
				"Объявление отклонено",
				msg,
				map[string]any{"listing_id": listing.ID, "reason": reason},
			)
		case "request_changes":
			msg := fmt.Sprintf("По вашему объявлению «%s» запрошены изменения.", listing.Name)
			if reason != "" {
				msg += " Комментарий: " + reason
			}
			_, _ = s.Notifications.CreateNotification(
				ctx,
				listing.HostID,
				notifications.TypeListingRejected,
				"Требуются изменения в объявлении",
				msg,
				map[string]any{"listing_id": listing.ID, "reason": reason},
			)
		}
	}

	return &listing, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// User Enforcement
// ─────────────────────────────────────────────────────────────────────────────

func (s *AdminService) GetUsers(ctx context.Context, role, status string) ([]db.User, error) {
	var users []db.User
	query := s.DB.WithContext(ctx).Order("created_at DESC")
	if role != "" {
		query = query.Where("role = ?", role)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (s *AdminService) EnforceUser(ctx context.Context, admin *db.User, targetUserID uuid.UUID, enforcementType, reason, note string, durationDays int) (*db.User, error) {
	var user db.User
	if err := s.DB.WithContext(ctx).Where("id = ?", targetUserID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	oldStatus := string(user.Status)
	newStatus := oldStatus

	var expiresAt *time.Time
	if durationDays > 0 {
		exp := time.Now().AddDate(0, 0, durationDays)
		expiresAt = &exp
	}

	switch enforcementType {
	case "warning":
		// User remains active, warning is logged
	case "temporary_restriction":
		newStatus = "suspended"
		user.Status = db.UserStatusSuspended
	case "temporary_block":
		newStatus = "suspended"
		user.Status = db.UserStatusSuspended
	case "permanent_block":
		newStatus = "banned"
		user.Status = db.UserStatusBanned
	case "unblock":
		newStatus = "active"
		user.Status = db.UserStatusActive
	default:
		return nil, fmt.Errorf("invalid enforcement type: %s", enforcementType)
	}

	if err := s.DB.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, err
	}

	// Deactivate previous active restrictions if unblocking
	if enforcementType == "unblock" {
		s.DB.WithContext(ctx).Model(&db.UserRestriction{}).Where("user_id = ?", targetUserID).Update("is_active", false)
	} else {
		restriction := &db.UserRestriction{
			ID:              uuid.New(),
			UserID:          targetUserID,
			AdminID:         admin.ID,
			EnforcementType: enforcementType,
			Reason:          reason,
			Note:            note,
			ExpiresAt:       expiresAt,
			IsActive:        true,
			CreatedAt:       time.Now(),
		}
		s.DB.WithContext(ctx).Create(restriction)
	}

	adminName := "Admin"
	if admin.FirstName != nil {
		adminName = *admin.FirstName
	}

	_ = s.CreateAuditLog(ctx, admin.ID, adminName, "user_"+enforcementType, "user", targetUserID.String(), reason, note, oldStatus, newStatus)

	if s.Notifications != nil {
		title := "Дисциплинарная мера"
		message := fmt.Sprintf("В отношении вашего профиля применено действие: %s. Причина: %s", enforcementType, reason)
		if enforcementType == "unblock" {
			title = "Ограничения сняты"
			message = "Все временные ограничения с вашего профиля были успешно сняты."
		}
		_, _ = s.Notifications.CreateNotification(
			ctx,
			targetUserID,
			notifications.TypeEnforcementIssued,
			title,
			message,
			map[string]any{
				"enforcement_type": enforcementType,
				"reason":           reason,
				"duration_days":    durationDays,
			},
		)
	}

	return &user, nil
}

func (s *AdminService) UpdateUserRole(ctx context.Context, admin *db.User, targetUserID uuid.UUID, newRole string) (*db.User, error) {
	var user db.User
	if err := s.DB.WithContext(ctx).Where("id = ?", targetUserID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	oldRole := string(user.Role)
	user.Role = db.UserRole(newRole)
	if err := s.DB.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, err
	}

	adminName := "Admin"
	if admin.FirstName != nil {
		adminName = *admin.FirstName
	}

	_ = s.CreateAuditLog(ctx, admin.ID, adminName, "user_role_change", "user", targetUserID.String(), "Role change by admin", newRole, oldRole, newRole)

	return &user, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Verification Requests
// ─────────────────────────────────────────────────────────────────────────────

func (s *AdminService) GetVerifications(ctx context.Context, status string) ([]db.VerificationRequest, error) {
	var items []db.VerificationRequest
	query := s.DB.WithContext(ctx).Order("created_at DESC")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *AdminService) ReviewVerification(ctx context.Context, admin *db.User, reqID uuid.UUID, action, reason, note string) (*db.VerificationRequest, error) {
	var v db.VerificationRequest
	if err := s.DB.WithContext(ctx).Where("id = ?", reqID).First(&v).Error; err != nil {
		return nil, fmt.Errorf("verification request not found: %w", err)
	}

	oldStatus := v.Status
	newStatus := oldStatus

	switch action {
	case "approve":
		newStatus = "approved"
	case "reject":
		newStatus = "rejected"
	case "request_changes":
		newStatus = "changes_requested"
	default:
		return nil, fmt.Errorf("invalid verification action: %s", action)
	}

	v.Status = newStatus
	v.RejectionReason = reason
	v.AdminNote = note
	if err := s.DB.WithContext(ctx).Save(&v).Error; err != nil {
		return nil, err
	}

	adminName := "Admin"
	if admin.FirstName != nil {
		adminName = *admin.FirstName
	}

	_ = s.CreateAuditLog(ctx, admin.ID, adminName, "verification_"+action, "verification", reqID.String(), reason, note, oldStatus, newStatus)

	if s.Notifications != nil {
		switch action {
		case "approve":
			_, _ = s.Notifications.CreateNotification(
				ctx,
				v.UserID,
				"verification_approved",
				"Верификация подтверждена",
				"Ваша заявка на верификацию успешно одобрена администрацией.",
				map[string]any{"verification_id": v.ID},
			)
		case "reject":
			msg := "Ваша заявка на верификацию отклонена."
			if reason != "" {
				msg += " Причина: " + reason
			}
			_, _ = s.Notifications.CreateNotification(
				ctx,
				v.UserID,
				"verification_rejected",
				"Верификация отклонена",
				msg,
				map[string]any{"verification_id": v.ID, "reason": reason},
			)
		case "request_changes":
			msg := "По вашей заявке на верификацию запрошены уточнения."
			if reason != "" {
				msg += " Комментарий: " + reason
			}
			_, _ = s.Notifications.CreateNotification(
				ctx,
				v.UserID,
				"verification_changes_requested",
				"Требуются изменения в заявке на верификацию",
				msg,
				map[string]any{"verification_id": v.ID, "reason": reason},
			)
		}
	}

	return &v, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Reports
// ─────────────────────────────────────────────────────────────────────────────

func (s *AdminService) GetReports(ctx context.Context, status string) ([]db.Report, error) {
	var reports []db.Report
	query := s.DB.WithContext(ctx).Order("created_at DESC")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Find(&reports).Error; err != nil {
		return nil, err
	}
	return reports, nil
}

func (s *AdminService) ResolveReport(ctx context.Context, admin *db.User, reportID uuid.UUID, action, reason, note string) (*db.Report, error) {
	var r db.Report
	if err := s.DB.WithContext(ctx).Where("id = ?", reportID).First(&r).Error; err != nil {
		return nil, fmt.Errorf("report not found: %w", err)
	}

	oldStatus := r.Status
	newStatus := "resolved"
	if action == "dismiss" {
		newStatus = "dismissed"
	}

	r.Status = newStatus
	r.ResolutionReason = reason
	r.ResolutionNote = note
	r.ResolverAdminID = &admin.ID
	if err := s.DB.WithContext(ctx).Save(&r).Error; err != nil {
		return nil, err
	}

	adminName := "Admin"
	if admin.FirstName != nil {
		adminName = *admin.FirstName
	}

	_ = s.CreateAuditLog(ctx, admin.ID, adminName, "report_"+action, "report", reportID.String(), reason, note, oldStatus, newStatus)

	return &r, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Audit Logs
// ─────────────────────────────────────────────────────────────────────────────

func (s *AdminService) GetAuditLogs(ctx context.Context) ([]db.AuditLog, error) {
	var logs []db.AuditLog
	if err := s.DB.WithContext(ctx).Order("created_at DESC").Limit(100).Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
