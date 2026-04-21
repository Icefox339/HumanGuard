package storage

import (
	"context"
	"encoding/json"
)

func (s *storage) GetSiteSettings(ctx context.Context, siteID string) (*ModuleSettings, error) {
	query := `SELECT settings FROM sites WHERE id = $1`

	var settingsJSON []byte
	err := s.db.QueryRowContext(ctx, query, siteID).Scan(&settingsJSON)
	if err != nil || len(settingsJSON) == 0 {
		return getDefaultSettings(), nil
	}

	var settings ModuleSettings
	json.Unmarshal(settingsJSON, &settings)
	return &settings, nil
}

func (s *storage) UpdateSiteSettings(ctx context.Context, siteID string, settings *ModuleSettings) error {
	settingsJSON, _ := json.Marshal(settings)
	query := `UPDATE sites SET settings = $1, updated_at = NOW() WHERE id = $2`
	_, err := s.db.ExecContext(ctx, query, settingsJSON, siteID)
	return err
}

func getDefaultSettings() *ModuleSettings {
	return &ModuleSettings{
		Collector: CollectorSettings{
			Enabled:       true,
			MouseTracking: true,
			ClickTracking: true,
		},
		Analyzer: AnalyzerSettings{
			Enabled:           true,
			RateLimiting:      true,
			HeadlessDetection: true,
			Thresholds: AnalyzerThreshold{
				Low:    30,
				Medium: 60,
				High:   80,
			},
		},
		Reaction: ReactionSettings{
			Enabled:          true,
			MediumRiskAction: "captcha",
			HighRiskAction:   "block",
			BlockDuration:    60,
		},
	}
}