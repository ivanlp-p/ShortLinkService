package cookie

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

type AuthCookie struct {
}

const (
	authCookieName = "auth_token"
	authSecretKey  = "your_secret_key" // хранить в конфиге
)

func GenerateUserID() string {
	return uuid.Must(uuid.NewV6()).String()
}

func sign(value string) string {
	h := hmac.New(sha256.New, []byte(authSecretKey))
	h.Write([]byte(value))
	return hex.EncodeToString(h.Sum(nil))
}

func MakeSignedCookie(userID string) *http.Cookie {
	sig := sign(userID)
	return &http.Cookie{
		Name:  authCookieName,
		Value: userID + "." + sig,
		Path:  "/",
	}
}

func ValidateAuthCookie(r *http.Request) (string, bool) {
	c, err := r.Cookie(authCookieName)
	if err != nil {
		return "", false
	}
	parts := strings.Split(c.Value, ".")
	if len(parts) != 2 {
		return "", false
	}
	userID, sig := parts[0], parts[1]
	if hmac.Equal([]byte(sig), []byte(sign(userID))) {
		return userID, true
	}
	return "", false
}
