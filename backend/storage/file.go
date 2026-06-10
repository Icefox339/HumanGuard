package storage

import (
	"context"
)

func (s *storage) CreateFile(ctx context.Context, file *FileRecord) error {
	if file.ID == "" {
		file.ID = generateID()
	}

	query := `
		INSERT INTO files (id, user_id, name, original_name, size, mime_type, hash, path, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := s.db.ExecContext(ctx, query,
		file.ID, file.UserID, file.Name, file.OriginalName,
		file.Size, file.MimeType, file.Hash, file.Path, file.CreatedAt,
	)
	return err
}

func (s *storage) GetFile(ctx context.Context, id string) (*FileRecord, error) {
	query := `SELECT id, user_id, name, original_name, size, mime_type, hash, path, created_at FROM files WHERE id = $1`

	var file FileRecord
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&file.ID, &file.UserID, &file.Name, &file.OriginalName,
		&file.Size, &file.MimeType, &file.Hash, &file.Path, &file.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (s *storage) DeleteFile(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM files WHERE id = $1`, id)
	return err
}

func (s *storage) ListUserFiles(ctx context.Context, userID string) ([]*FileRecord, error) {
	query := `SELECT id, user_id, name, original_name, size, mime_type, hash, path, created_at FROM files WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var files []*FileRecord
	for rows.Next() {
		var f FileRecord
		if err := rows.Scan(&f.ID, &f.UserID, &f.Name, &f.OriginalName, &f.Size, &f.MimeType, &f.Hash, &f.Path, &f.CreatedAt); err != nil {
			return nil, err
		}
		files = append(files, &f)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}
