package reaction

import (
	"context"
	"fmt"
	"log"
	"time"
)

type RiskLevel int

const (
	RiskLow    RiskLevel = 0
	RiskMedium RiskLevel = 1
	RiskHigh   RiskLevel = 2
)

type ReactionConfig struct {
	Enabled                  bool   `json:"enabled"`
	LowRiskAction            string `json:"low_risk_action"`
	MediumRiskAction         string `json:"medium_risk_action"`
	HighRiskAction           string `json:"high_risk_action"`
	BlockDurationMinutes     int    `json:"block_duration_minutes"`
	BlockDurationPermanent   bool   `json:"block_duration_permanent"`
	BlockMessage             string `json:"block_message"`
	AddToBlacklist           bool   `json:"add_to_blacklist"`
	BlacklistDurationMinutes int    `json:"blacklist_duration_minutes"`
	CaptchaProvider          string `json:"captcha_provider"`
	CaptchaSitekey           string `json:"captcha_sitekey"`
}

func DefaultReactionConfig() ReactionConfig {
	return ReactionConfig{
		Enabled:                  true,
		LowRiskAction:            "allow",
		MediumRiskAction:         "captcha",
		HighRiskAction:           "block",
		BlockDurationMinutes:     60,
		BlockDurationPermanent:   false,
		BlockMessage:             "Access denied. Your activity appears to be automated.",
		AddToBlacklist:           true,
		BlacklistDurationMinutes: 1440,
		CaptchaProvider:          "hcaptcha",
		CaptchaSitekey:           "10000000-ffff-ffff-ffff-000000000001",
	}
}

type SessionInfo struct {
	ID           string `json:"id"`
	SiteID       string `json:"site_id"`
	IP           string `json:"ip"`
	RiskScore    int    `json:"risk_score"`
	IsBlocked    bool   `json:"is_blocked"`
	CaptchaShown bool   `json:"captcha_shown"`
}

type ReactionResult struct {
	Action         string `json:"action"`
	Blocked        bool   `json:"blocked"`
	Captcha        bool   `json:"captcha"`
	Message        string `json:"message,omitempty"`
	Blacklist      bool   `json:"blacklist"`
	CaptchaSitekey string `json:"captcha_sitekey,omitempty"`
}

type Store interface {
	BlockSession(ctx context.Context, id string) error
	UnblockSession(ctx context.Context, id string) error
	MarkCaptchaShown(ctx context.Context, id string) error
	AddToBlacklist(ctx context.Context, siteID, ip, reason string, duration time.Duration) error
	RemoveFromBlacklist(ctx context.Context, siteID, ip string) error
	IsBlacklisted(ctx context.Context, siteID, ip string) (bool, error)
}

type Service struct {
	store  Store
	config ReactionConfig
}

func NewService(store Store, config ReactionConfig) *Service {
	return &Service{
		store:  store,
		config: config,
	}
}

func (s *Service) React(ctx context.Context, session *SessionInfo) ReactionResult {
	if !s.config.Enabled {
		return ReactionResult{Action: "allow"}
	}

	if session.IsBlocked {
		return ReactionResult{
			Action:  "block",
			Blocked: true,
			Message: s.config.BlockMessage,
		}
	}

	switch s.getRiskLevel(session.RiskScore) {
	case RiskHigh:
		return s.handleHighRisk(ctx, session)
	case RiskMedium:
		return s.handleMediumRisk(ctx, session)
	default:
		return s.handleLowRisk()
	}
}

func (s *Service) getRiskLevel(score int) RiskLevel {
	switch {
	case score >= 80:
		return RiskHigh
	case score >= 50:
		return RiskMedium
	default:
		return RiskLow
	}
}

func (s *Service) handleLowRisk() ReactionResult {
	return ReactionResult{Action: s.config.LowRiskAction}
}

func (s *Service) handleMediumRisk(ctx context.Context, session *SessionInfo) ReactionResult {
	switch s.config.MediumRiskAction {
	case "captcha":
		if err := s.store.MarkCaptchaShown(ctx, session.ID); err != nil {
			log.Printf("Failed to mark captcha shown for session %s: %v", session.ID, err)
		}
		return ReactionResult{
			Action:         "captcha",
			Captcha:        true,
			CaptchaSitekey: s.config.CaptchaSitekey,
		}
	case "block":
		if err := s.store.BlockSession(ctx, session.ID); err != nil {
			log.Printf("Failed to block session %s: %v", session.ID, err)
		}
		return ReactionResult{
			Action:  "block",
			Blocked: true,
			Message: s.config.BlockMessage,
		}
	default:
		return ReactionResult{Action: "allow"}
	}
}

func (s *Service) handleHighRisk(ctx context.Context, session *SessionInfo) ReactionResult {
	if err := s.store.BlockSession(ctx, session.ID); err != nil {
		log.Printf("Failed to block session %s: %v", session.ID, err)
	}

	if s.config.AddToBlacklist {
		var duration time.Duration
		if !s.config.BlockDurationPermanent {
			duration = time.Duration(s.config.BlacklistDurationMinutes) * time.Minute
		}
		if err := s.store.AddToBlacklist(ctx, session.SiteID, session.IP, fmt.Sprintf("High risk score: %d", session.RiskScore), duration); err != nil {
			log.Printf("Failed to add to blacklist session %s ip %s: %v", session.ID, session.IP, err)
		}
	}

	log.Printf("BLOCKED: session=%s ip=%s risk=%d", session.ID, session.IP, session.RiskScore)

	return ReactionResult{
		Action:    "block",
		Blocked:   true,
		Message:   s.config.BlockMessage,
		Blacklist: s.config.AddToBlacklist,
	}
}
