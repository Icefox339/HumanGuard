package auth

import (
	"sync"
	"time"
)

type UserSession struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	LastSeen  time.Time `json:"last_seen"`
	ExpiresAt time.Time `json:"expires_at"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
}

type UserSessionManager struct {
	sessions sync.Map
	ttl      time.Duration
}

func NewUserSessionManager(ttl time.Duration) *UserSessionManager {
	m := &UserSessionManager{ttl: ttl}
	go m.cleanupLoop()
	return m
}

func (m *UserSessionManager) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		now := time.Now()
		m.sessions.Range(func(key, value interface{}) bool {
			sess := value.(*UserSession)
			if now.After(sess.ExpiresAt) {
				m.sessions.Delete(key)
			}
			return true
		})
	}
}

func (m *UserSessionManager) Create(sessionID, userID, email, role, ip, userAgent string) *UserSession {
	now := time.Now()
	sess := &UserSession{
		ID:        sessionID,
		UserID:    userID,
		Email:     email,
		Role:      role,
		CreatedAt: now,
		LastSeen:  now,
		ExpiresAt: now.Add(m.ttl),
		IP:        ip,
		UserAgent: userAgent,
	}
	m.sessions.Store(sessionID, sess)
	return sess
}

func (m *UserSessionManager) Get(sessionID string) (*UserSession, bool) {
	val, ok := m.sessions.Load(sessionID)
	if !ok {
		return nil, false
	}
	sess := val.(*UserSession)
	if time.Now().After(sess.ExpiresAt) {
		m.sessions.Delete(sessionID)
		return nil, false
	}
	return sess, true
}

func (m *UserSessionManager) UpdateLastSeen(sessionID string) {
	if val, ok := m.sessions.Load(sessionID); ok {
		sess := val.(*UserSession)
		sess.LastSeen = time.Now()
		sess.ExpiresAt = time.Now().Add(m.ttl)
	}
}

func (m *UserSessionManager) Delete(sessionID string) {
	m.sessions.Delete(sessionID)
}

func (m *UserSessionManager) ListAll() []*UserSession {
	var result []*UserSession
	m.sessions.Range(func(key, value interface{}) bool {
		result = append(result, value.(*UserSession))
		return true
	})
	return result
}