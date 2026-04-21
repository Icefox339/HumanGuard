package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (s *storage) CreateUser(ctx context.Context, user *User) error {
	if user.ID == "" {
		user.ID = generateID()
	}
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	if user.Role == "" {
		user.Role = "user"
	}

	query := `
		INSERT INTO users (
			id, email, name, avatar_url, role, 
			totp_secret, password_hash, oauth_provider, oauth_id, 
			created_at, updated_at, last_login
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11,
			$12, $13, $14
		)
	`

	_, err := s.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.AvatarURL,
		user.Role,
		user.TOTPSecret,
		user.PasswordHash,
		user.OAuthProvider,
		user.OAuthID,
		user.CreatedAt,
		user.UpdatedAt,
		user.LastLogin,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (s *storage) GetUserByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT 
			id, email, name, avatar_url, role,
			totp_secret, password_hash,
			oauth_provider, oauth_id,
			created_at, updated_at, last_login
		FROM users 
		WHERE id = $1
	`

	var user User

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.AvatarURL,
		&user.Role,
		&user.TOTPSecret,
		&user.PasswordHash,
		&user.OAuthProvider,
		&user.OAuthID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLogin,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (s *storage) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT 
			id, email, name, avatar_url, role,
			totp_secret, password_hash,
			oauth_provider, oauth_id,
			created_at, updated_at, last_login
		FROM users 
		WHERE email = $1
	`

	var user User

	err := s.db.QueryRowContext(ctx, query, strings.ToLower(email)).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.AvatarURL,
		&user.Role,
		&user.TOTPSecret,
		&user.PasswordHash,
		&user.OAuthProvider,
		&user.OAuthID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLogin,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

func (s *storage) GetUserByOAuth(ctx context.Context, provider, oauthID string) (*User, error) {
	query := `
		SELECT 
			id, email, name, avatar_url, role,
			totp_secret, password_hash,
			oauth_provider, oauth_id,
			created_at, updated_at, last_login
		FROM users 
		WHERE oauth_provider = $1 AND oauth_id = $2
	`

	var user User

	err := s.db.QueryRowContext(ctx, query, provider, oauthID).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.AvatarURL,
		&user.Role,
		&user.TOTPSecret,
		&user.PasswordHash,
		&user.OAuthProvider,
		&user.OAuthID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLogin,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by oauth: %w", err)
	}

	return &user, nil
}

func (s *storage) UpdateUser(ctx context.Context, user *User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users 
		SET 
			name = $1,
			role = $2,
			updated_at = $3,
			last_login = $4
		WHERE id = $5
	`

	result, err := s.db.ExecContext(ctx, query,
		user.Name,
		user.Role,
		user.UpdatedAt,
		user.LastLogin,
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (s *storage) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	query := `
		UPDATE users 
		SET 
			password_hash = $1,
			updated_at = $2
		WHERE id = $3
	`

	result, err := s.db.ExecContext(ctx, query, passwordHash, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (s *storage) UpdateAvatar(ctx context.Context, userID, avatarURL string) error {
	query := `
		UPDATE users 
		SET 
			avatar_url = $1,
			updated_at = $2
		WHERE id = $3
	`

	result, err := s.db.ExecContext(ctx, query, avatarURL, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update avatar: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (s *storage) DeleteUser(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *storage) UpdateLastLogin(ctx context.Context, userID string) error {
	now := time.Now()
	query := `
		UPDATE users 
		SET 
			last_login = $1,
			updated_at = $1
		WHERE id = $2
	`

	_, err := s.db.ExecContext(ctx, query, now, userID)
	return err
}

func (s *storage) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := s.db.QueryRowContext(ctx, query, strings.ToLower(email)).Scan(&exists)
	return exists, err
}

