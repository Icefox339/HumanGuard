package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService struct {
	secret []byte
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{secret: []byte(secret)}
}

func (j *JWTService) GenerateToken(userID, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTService) GenerateTokenWithSessionID(userID, role, sessionID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"sid":     sessionID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTService) ValidateToken(tokenString string) (string, string, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})
	if err != nil {
		return "", "", "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", "", fmt.Errorf("invalid token")
	}
	userID, _ := claims["user_id"].(string)
	role, _ := claims["role"].(string)
	sessionID, _ := claims["sid"].(string)

	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	return userID, role, sessionID, nil
}

type TOTPService struct{}

func NewTOTPService() *TOTPService {
	return &TOTPService{}
}

func (t *TOTPService) GenerateSecret() (string, error) {
	bytes := make([]byte, 20)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(bytes), nil
}

func (t *TOTPService) GenerateCode(secret string) string {
	secret = strings.ToUpper(strings.TrimSpace(secret))
	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return ""
	}

	counter := time.Now().Unix() / 30
	msg := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		msg[i] = byte(counter)
		counter >>= 8
	}

	h := hmac.New(sha1.New, key)
	h.Write(msg)
	hash := h.Sum(nil)

	offset := hash[len(hash)-1] & 0x0f
	binary := (int(hash[offset])&0x7f)<<24 |
		(int(hash[offset+1])&0xff)<<16 |
		(int(hash[offset+2])&0xff)<<8 |
		(int(hash[offset+3]) & 0xff)

	return fmt.Sprintf("%06d", binary%1000000)
}

func (t *TOTPService) ValidateCode(secret, code string) bool {
	serverCode := t.GenerateCode(secret)
	fmt.Printf("DEBUG: Server generated code: %s, Client provided: %s\n", serverCode, code)
	return serverCode == code
}

func (t *TOTPService) GenerateQRURL(email, secret string) string {
	return fmt.Sprintf("otpauth://totp/HumanGuard:%s?secret=%s&issuer=HumanGuard", email, secret)
}
