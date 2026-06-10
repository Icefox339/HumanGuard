package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
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
			password_hash, is_verified, totp_secret,
			created_at, updated_at, last_login
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11
		)
	`

	_, err := s.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.AvatarURL,
		user.Role,
		user.PasswordHash,
		user.IsVerified,
		user.TOTPSecret,
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

func (s *storage) ListUsers(ctx context.Context) ([]*User, error) {
	query := `
		SELECT
			id, email, name, avatar_url, role,
			totp_secret, password_hash, is_verified,
			created_at, updated_at, last_login
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer func() { _ = rows.Close() }()

	users := make([]*User, 0)
	for rows.Next() {
		var user User
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.AvatarURL,
			&user.Role,
			&user.TOTPSecret,
			&user.PasswordHash,
			&user.IsVerified,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.LastLogin,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, nil
}

func (s *storage) GetUserByID(ctx context.Context, id string) (*User, error) {
	query := `
		SELECT
			id, email, name, avatar_url, role,
			totp_secret, password_hash, is_verified,
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
		&user.IsVerified,
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
			totp_secret, password_hash, is_verified,
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
		&user.IsVerified,
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

func (s *storage) UpdateUser(ctx context.Context, user *User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET
			name = $1,
			role = $2,
			avatar_url = $3,
			updated_at = $4,
			last_login = $5
		WHERE id = $6
	`

	result, err := s.db.ExecContext(ctx, query,
		user.Name,
		user.Role,
		user.AvatarURL,
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
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("Failed to rollback transaction: %v", err)
		}
	}()

	_, err = tx.ExecContext(ctx, `DELETE FROM user_oauth WHERE user_id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete user oauth: %w", err)
	}

	result, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
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

// ========== OAUTH STORAGE ==========

func (s *storage) AddUserOAuth(ctx context.Context, userID, provider, oauthID string) error {
	query := `
		INSERT INTO user_oauth (id, user_id, provider, oauth_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (provider, oauth_id) DO NOTHING
	`

	_, err := s.db.ExecContext(ctx, query,
		generateID(), userID, provider, oauthID, time.Now())

	return err
}

func (s *storage) GetUserByOAuth(ctx context.Context, provider, oauthID string) (*User, error) {
	query := `
		SELECT u.id, u.email, u.name, u.avatar_url, u.role, 
		       u.password_hash, u.is_verified, u.totp_secret,
		       u.created_at, u.updated_at, u.last_login
		FROM users u
		JOIN user_oauth o ON u.id = o.user_id
		WHERE o.provider = $1 AND o.oauth_id = $2
	`

	var user User
	err := s.db.QueryRowContext(ctx, query, provider, oauthID).Scan(
		&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.Role,
		&user.PasswordHash, &user.IsVerified, &user.TOTPSecret,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by oauth: %w", err)
	}

	return &user, nil
}

func (s *storage) GetUserOAuths(ctx context.Context, userID string) ([]*UserOAuth, error) {
	query := `SELECT id, user_id, provider, oauth_id, created_at FROM user_oauth WHERE user_id = $1 ORDER BY created_at`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var oauths []*UserOAuth
	for rows.Next() {
		var o UserOAuth
		if err := rows.Scan(&o.ID, &o.UserID, &o.Provider, &o.OAuthID, &o.CreatedAt); err != nil {
			return nil, err
		}
		oauths = append(oauths, &o)
	}

	return oauths, nil
}

func (s *storage) RemoveUserOAuth(ctx context.Context, userID, provider string) error {
	query := `DELETE FROM user_oauth WHERE user_id = $1 AND provider = $2`
	_, err := s.db.ExecContext(ctx, query, userID, provider)
	return err
}

func (s *storage) GetOrCreateUserByOAuth(ctx context.Context, provider, oauthID, email, name string) (*User, error) {
	user, err := s.GetUserByOAuth(ctx, provider, oauthID)
	if err == nil {
		return user, nil
	}

	user, err = s.GetUserByEmail(ctx, email)
	if err == nil {
		if err := s.AddUserOAuth(ctx, user.ID, provider, oauthID); err != nil {
			log.Printf("Failed to link OAuth %s to user %s: %v", provider, user.ID, err)
		}
		return user, nil
	}

	newUser := &User{
		Email:      email,
		Name:       name,
		Role:       "user",
		IsVerified: true,
	}

	if err := s.CreateUser(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to create user from OAuth: %w", err)
	}

	if err := s.AddUserOAuth(ctx, newUser.ID, provider, oauthID); err != nil {
		log.Printf("Failed to add OAuth for new user: %v", err)
	}

	return newUser, nil
}
