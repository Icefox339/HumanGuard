package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"time"
)

type Storage interface {
	UserStorage
	SessionStorage
	FileStorage
	ShareStorage
	SiteStorage
	SettingsStorage
	BlacklistStorage
	AccessLogStorage

	Close() error
	Ping() error
}

type Config struct {
	DBURL       string
	UploadDir   string
	MaxFileSize int64
}

type User struct {
	ID            string     `json:"id"`
	Email         string     `json:"email"`
	Name          string     `json:"name"`
	AvatarURL     *string    `json:"avatar_url"`
	Role          string     `json:"role"`
	TOTPSecret    *string    `json:"-"`
	PasswordHash  string     `json:"-"`
	IsVerified    bool       `json:"is_verified"`
	OAuthProvider *string    `json:"oauth_provider"`
	OAuthID       *string    `json:"-"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastLogin     *time.Time `json:"last_login"`
}
type Session struct {
	ID           string    `json:"id"`
	SiteID       *string   `json:"site_id"`
	IP           string    `json:"ip"`
	UserAgent    string    `json:"user_agent"`
	Device       string    `json:"device"`
	Location     string    `json:"location"`
	IsActive     bool      `json:"is_active"`
	RiskScore    int       `json:"risk_score"`
	IsBlocked    bool      `json:"is_blocked"`
	CaptchaShown bool      `json:"captcha_shown"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type ModuleSettings struct {
	Collector CollectorSettings `json:"collector"`
	Analyzer  AnalyzerSettings  `json:"analyzer"`
	Reaction  ReactionSettings  `json:"reaction"`
}

type CollectorSettings struct {
	Enabled            bool `json:"enabled"`
	MouseTracking      bool `json:"mouse_tracking"`
	ClickTracking      bool `json:"click_tracking"`
	ScrollTracking     bool `json:"scroll_tracking"`
	KeystrokeTracking  bool `json:"keystroke_tracking"`
	FingerprintEnabled bool `json:"fingerprint_enabled"`
}

type AnalyzerSettings struct {
	Enabled           bool              `json:"enabled"`
	RateLimiting      bool              `json:"rate_limiting"`
	PatternAnalysis   bool              `json:"pattern_analysis"`
	HeadlessDetection bool              `json:"headless_detection"`
	Thresholds        AnalyzerThreshold `json:"thresholds"`
}

type AnalyzerThreshold struct {
	Low    int `json:"low"`
	Medium int `json:"medium"`
	High   int `json:"high"`
}

type ReactionSettings struct {
	Enabled          bool   `json:"enabled"`
	LowRiskAction    string `json:"low_risk_action"`    // allow, log
	MediumRiskAction string `json:"medium_risk_action"` // allow, captcha, challenge
	HighRiskAction   string `json:"high_risk_action"`   // block, captcha, redirect
	BlockDuration    int    `json:"block_duration"`     // минуты
	CaptchaProvider  string `json:"captcha_provider"`   // hcaptcha, recaptcha
}

type Site struct {
	ID           string          `json:"id"`
	UserID       string          `json:"user_id"`
	Name         string          `json:"name"`
	Domain       string          `json:"domain"`
	OriginServer string          `json:"origin_server"`
	Status       string          `json:"status"`
	Settings     *ModuleSettings `json:"settings"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type BlacklistEntry struct {
	ID        string     `json:"id"`
	SiteID    string     `json:"site_id"`
	IP        string     `json:"ip"`
	Reason    string     `json:"reason"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type AccessLog struct {
	ID         string    `json:"id"`
	SessionID  string    `json:"session_id"`
	SiteID     string    `json:"site_id"`
	IP         string    `json:"ip"`
	Path       string    `json:"path"`
	Method     string    `json:"method"`
	UserAgent  string    `json:"user_agent"`
	Referer    string    `json:"referer"`
	StatusCode int       `json:"status_code"`
	RiskScore  int       `json:"risk_score"`
	Action     string    `json:"action"`
	CreatedAt  time.Time `json:"created_at"`
}

type LogStats struct {
	TotalRequests   int64   `json:"total_requests"`
	BlockedRequests int64   `json:"blocked_requests"`
	CaptchaShown    int64   `json:"captcha_shown"`
	AllowedRequests int64   `json:"allowed_requests"`
	UniqueIPs       int64   `json:"unique_ips"`
	AvgRiskScore    float64 `json:"avg_risk_score"`
}

type FileRecord struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Name         string    `json:"name"`
	OriginalName string    `json:"original_name"`
	Size         int64     `json:"size"`
	MimeType     string    `json:"mime_type"`
	Hash         string    `json:"hash"`
	Path         string    `json:"path"`
	CreatedAt    time.Time `json:"created_at"`
}

type ShareRecord struct {
	ID        string    `json:"id"`
	FileID    string    `json:"file_id"`
	Token     string    `json:"token"`
	SharedBy  string    `json:"shared_by"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type FileStorage interface {
	CreateFile(ctx context.Context, file *FileRecord) error
	GetFile(ctx context.Context, id string) (*FileRecord, error)
	DeleteFile(ctx context.Context, id string) error
	ListUserFiles(ctx context.Context, userID string) ([]*FileRecord, error)
}

type ShareStorage interface {
	CreateShare(ctx context.Context, share *ShareRecord) (string, error)
	GetFileByShareToken(ctx context.Context, token string) (*FileRecord, error)
}
type AccessLogStorage interface {
	LogAccess(ctx context.Context, log *AccessLog) error
	GetAccessLogs(ctx context.Context, siteID string) ([]*AccessLog, error)
	GetLogStats(ctx context.Context, siteID string, from, to time.Time) (*LogStats, error)
	CleanupOldLogs(ctx context.Context, siteID string, before time.Time) (int64, error)
}

type BlacklistStorage interface {
	AddToBlacklist(ctx context.Context, entry *BlacklistEntry) error
	RemoveFromBlacklist(ctx context.Context, siteID, ip string) error
	IsBlacklisted(ctx context.Context, siteID, ip string) (bool, error)
	ListBlacklist(ctx context.Context, siteID string) ([]*BlacklistEntry, error)
}

type UserStorage interface {
	CreateUser(ctx context.Context, user *User) error
	ListUsers(ctx context.Context) ([]*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id string) error
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByOAuth(ctx context.Context, provider, oauthID string) (*User, error)
	GetOrCreateUserByOAuth(ctx context.Context, provider, oauthID, email, name string) (*User, error)
	UpdateLastLogin(ctx context.Context, userID string) error
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	UpdateAvatar(ctx context.Context, userID, avatarURL string) error
	CheckEmailExists(ctx context.Context, email string) (bool, error)
}

type SessionStorage interface {
	CreateSession(ctx context.Context, session *Session) error
	GetSession(ctx context.Context, id string) (*Session, error)
	GetSessionByCookie(ctx context.Context, cookie string) (*Session, error)
	UpdateSession(ctx context.Context, session *Session) error
	UpdateSessionActivity(ctx context.Context, id string) error
	DeactivateSession(ctx context.Context, id string) error

	GetActiveSessionsBySite(ctx context.Context, siteID string, limit int) ([]*Session, error)
	GetSuspiciousSessions(ctx context.Context, siteID string, minRisk int) ([]*Session, error)
	BlockSession(ctx context.Context, id string) error
	UnblockSession(ctx context.Context, id string) error
	UpdateRiskScore(ctx context.Context, id string, score int) error
	MarkCaptchaShown(ctx context.Context, id string) error
	CleanupExpiredSessions(ctx context.Context) (int64, error)
	GetSessionStats(ctx context.Context, siteID string) (*SessionStats, error)
}

type SiteStorage interface {
	CreateSite(ctx context.Context, site *Site) error
	GetSite(ctx context.Context, id string) (*Site, error)
	GetSiteByDomain(ctx context.Context, domain string) (*Site, error)
	UpdateSite(ctx context.Context, site *Site) error
	DeleteSite(ctx context.Context, id string) error
	UpdateSiteStatus(ctx context.Context, siteID, status string) error
	ActivateSite(ctx context.Context, siteID string) error
	SuspendSite(ctx context.Context, siteID string) error
	GetSitesByUserID(ctx context.Context, userID string) ([]*Site, error)
}

type SettingsStorage interface {
	GetSiteSettings(ctx context.Context, siteID string) (*ModuleSettings, error)
	UpdateSiteSettings(ctx context.Context, siteID string, settings *ModuleSettings) error
}

func generateID() string {
	return uuid.New().String()
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func isUniqueViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505"
	}
	return false
}
