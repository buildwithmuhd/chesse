package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type TokenManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	now        func() time.Time
}

type Claims struct {
	UserID int64  `json:"sub"`
	Email  string `json:"email"`
	Exp    int64  `json:"exp"`
}

func NewTokenManager(secret []byte, accessTTL, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{secret: secret, accessTTL: accessTTL, refreshTTL: refreshTTL, now: time.Now}
}

func (m *TokenManager) NewAccessToken(user User) (string, error) {
	if len(m.secret) < 32 {
		return "", errors.New("jwt secret must be at least 32 bytes")
	}
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	claims := Claims{UserID: user.ID, Email: user.Email, Exp: m.now().Add(m.accessTTL).Unix()}
	return m.sign(header, claims)
}

func (m *TokenManager) ParseAccessToken(token string) (Claims, error) {
	var claims Claims
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return claims, errors.New("invalid token")
	}
	body := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(parts[2]), []byte(m.signature(body))) {
		return claims, errors.New("invalid signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return claims, err
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return claims, err
	}
	if claims.UserID <= 0 || claims.Exp <= m.now().Unix() {
		return claims, errors.New("expired token")
	}
	return claims, nil
}

func (m *TokenManager) NewRefreshToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func (m *TokenManager) RefreshExpiry() time.Time {
	return m.now().Add(m.refreshTTL)
}

func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func NewID() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", raw[0:4], raw[4:6], raw[6:8], raw[8:10], raw[10:]), nil
}

func (m *TokenManager) sign(header map[string]string, claims Claims) (string, error) {
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	body := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(claimsJSON)
	return body + "." + m.signature(body), nil
}

func (m *TokenManager) signature(body string) string {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(body))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func bearerToken(header string) (string, bool) {
	prefix, token, ok := strings.Cut(header, " ")
	return token, ok && strings.EqualFold(prefix, "Bearer") && token != ""
}

func claimUserID(claims Claims) string {
	return strconv.FormatInt(claims.UserID, 10)
}
