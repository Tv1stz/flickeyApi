export interface AdminOverviewMetrics {
	total_listings: number;
	pending_listings: number;
	approved_listings: number;
	rejected_listings: number;
	total_users: number;
	active_users: number;
	blocked_users: number;
	pending_verifications: number;
	open_reports: number;
}

export type ModerationAction = 'approve' | 'reject' | 'request_changes' | 'suspend';

export interface ModerationPayload {
	action: ModerationAction;
	reason?: string;
	note?: string;
}

export type EnforcementType =
	| 'warning'
	| 'temporary_restriction'
	| 'temporary_block'
	| 'permanent_block'
	| 'unblock';

export interface EnforcementPayload {
	enforcement_type: EnforcementType;
	reason: string;
	note?: string;
	duration_days?: number;
}

export type ProviderType = 'individual' | 'individual_entrepreneur' | 'self_employed' | 'legal_entity';

export type VerificationStatus = 'pending' | 'under_review' | 'approved' | 'rejected' | 'changes_requested';

export interface VerificationRequestItem {
	id: string;
	user_id: string;
	provider_type: ProviderType;
	legal_name: string;
	unp?: string;
	status: VerificationStatus;
	rejection_reason?: string;
	admin_note?: string;
	requested_changes?: string[];
	requisites?: Record<string, any>;
	documents?: Array<{ id: string; name: string; url: string; status?: string }>;
	created_at: string;
	updated_at?: string;
}

export type ReportStatus = 'open' | 'under_review' | 'resolved' | 'dismissed';

export interface ReportItem {
	id: string;
	reporter_id: string;
	target_type: 'listing' | 'user';
	target_id: string;
	reason: string;
	description?: string;
	status: ReportStatus;
	resolution_reason?: string;
	resolution_note?: string;
	resolver_admin_id?: string;
	created_at: string;
}

export interface AuditLogEntry {
	id: string;
	admin_id: string;
	admin_name: string;
	action: string;
	target_type: string;
	target_id: string;
	reason?: string;
	note?: string;
	old_status?: string;
	new_status?: string;
	created_at: string;
}

export interface UserRestrictionRecord {
	id: string;
	user_id: string;
	admin_id: string;
	enforcement_type: EnforcementType;
	reason: string;
	note?: string;
	expires_at?: string;
	is_active: boolean;
	created_at: string;
}
