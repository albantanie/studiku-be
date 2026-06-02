package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"studi-ku-backend/internal/models"
)

type authTokenPayload struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Exp   int64  `json:"exp"`
}

func issueAuthToken(user *models.LoginUser) (string, error) {
	payload := authTokenPayload{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
		Exp:   time.Now().Add(12 * time.Hour).Unix(),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	bodyPart := base64.RawURLEncoding.EncodeToString(body)
	signature := signTokenPart(bodyPart)
	return bodyPart + "." + signature, nil
}

func signTokenPart(bodyPart string) string {
	mac := hmac.New(sha256.New, []byte(authSecret()))
	mac.Write([]byte(bodyPart))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func authSecret() string {
	if value := os.Getenv("APP_SECRET"); strings.TrimSpace(value) != "" {
		return value
	}
	return "studiku-local-dev-secret"
}

func (h *Handler) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if strings.TrimSpace(token) == "" {
			fail(c, http.StatusUnauthorized, "authentication required")
			c.Abort()
			return
		}
		parts := strings.Split(token, ".")
		if len(parts) != 2 || !hmac.Equal([]byte(parts[1]), []byte(signTokenPart(parts[0]))) {
			fail(c, http.StatusUnauthorized, "invalid authentication token")
			c.Abort()
			return
		}
		raw, err := base64.RawURLEncoding.DecodeString(parts[0])
		if err != nil {
			fail(c, http.StatusUnauthorized, "invalid authentication token")
			c.Abort()
			return
		}
		var payload authTokenPayload
		if err := json.Unmarshal(raw, &payload); err != nil || payload.Exp < time.Now().Unix() {
			fail(c, http.StatusUnauthorized, "authentication token expired")
			c.Abort()
			return
		}
		if !h.repo.UserExists(payload.ID, payload.Email, payload.Role) {
			fail(c, http.StatusUnauthorized, "authentication user is invalid")
			c.Abort()
			return
		}
		c.Set("authUser", models.AuthUser{ID: payload.ID, Email: payload.Email, Role: payload.Role})
		c.Next()
	}
}

func requireRole(roles ...string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, role := range roles {
		allowed[role] = true
	}
	return func(c *gin.Context) {
		authUser, ok := currentAuthUser(c)
		if !ok || !allowed[authUser.Role] {
			fail(c, http.StatusForbidden, "permission denied")
			c.Abort()
			return
		}
		c.Next()
	}
}

func currentAuthUser(c *gin.Context) (models.AuthUser, bool) {
	value, ok := c.Get("authUser")
	if !ok {
		return models.AuthUser{}, false
	}
	authUser, ok := value.(models.AuthUser)
	return authUser, ok
}

func currentUserID(c *gin.Context) int {
	authUser, ok := currentAuthUser(c)
	if !ok {
		return 0
	}
	return authUser.ID
}

func intParam(c *gin.Context, key string) (int, error) {
	id, err := strconv.Atoi(c.Param(key))
	if err != nil {
		fail(c, http.StatusBadRequest, "invalid "+key)
		return 0, err
	}
	return id, nil
}
