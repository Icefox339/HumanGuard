package storage

import (
	"context"
)

func (s *storage) CreateShare(ctx context.Context, share *ShareRecord) (string, error) {
	share.ID = generateID()

	query := `
		INSERT INTO shares (id, file_id, token, shared_by, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := s.db.ExecContext(ctx, query,
		share.ID, share.FileID, share.Token, share.SharedBy, share.ExpiresAt, share.CreatedAt,
	)
	if err != nil {
		return "", err
	}

	return share.Token, nil
}

func (s *storage) GetFileByShareToken(ctx context.Context, token string) (*FileRecord, error) {
	query := `
		SELECT f.id, f.user_id, f.name, f.original_name, f.size, f.mime_type, f.hash, f.path, f.created_at
		FROM files f
		JOIN shares s ON f.id = s.file_id
		WHERE s.token = $1 AND (s.expires_at IS NULL OR s.expires_at < '1970-01-01' OR s.expires_at > NOW())
	`

	var file FileRecord
	err := s.db.QueryRowContext(ctx, query, token).Scan(
		&file.ID, &file.UserID, &file.Name, &file.OriginalName,
		&file.Size, &file.MimeType, &file.Hash, &file.Path, &file.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &file, nil
}