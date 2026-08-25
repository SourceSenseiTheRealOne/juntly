package authn

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/SourceSenseiTheRealOne/juntly/backend/internal/users"
)

type verifiedIdentityContextKey struct{}

type unauthorizedResponse struct {
	Error unauthorizedError `json:"error"`
}

type unauthorizedError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"requestId"`
}

func RequireVerifiedIdentity(verifier Verifier, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		token, ok := bearerToken(request.Header.Get("Authorization"))
		if !ok || verifier == nil {
			writeUnauthorized(w, request.Header.Get("X-Request-ID"))
			return
		}

		identity, err := verifier.Verify(request.Context(), token)
		if err != nil || strings.TrimSpace(identity.Subject) == "" {
			writeUnauthorized(w, request.Header.Get("X-Request-ID"))
			return
		}

		next.ServeHTTP(w, request.WithContext(context.WithValue(request.Context(), verifiedIdentityContextKey{}, identity)))
	})
}

func IdentityFromContext(ctx context.Context) (users.VerifiedIdentity, bool) {
	identity, ok := ctx.Value(verifiedIdentityContextKey{}).(users.VerifiedIdentity)
	return identity, ok
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func writeUnauthorized(w http.ResponseWriter, requestedID string) {
	requestID := requestIDFromHeader(requestedID)
	w.Header().Set("X-Request-ID", requestID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(unauthorizedResponse{Error: unauthorizedError{
		Code:      "UNAUTHORIZED",
		Message:   "Unauthorized",
		RequestID: requestID,
	}})
}

func requestIDFromHeader(value string) string {
	if validRequestID(value) {
		return value
	}

	var randomBytes [16]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return "req_unavailable"
	}

	return "req_" + hex.EncodeToString(randomBytes[:])
}

func validRequestID(value string) bool {
	if len(value) < 8 || len(value) > 128 {
		return false
	}

	for _, char := range value {
		switch {
		case char >= 'A' && char <= 'Z':
		case char >= 'a' && char <= 'z':
		case char >= '0' && char <= '9':
		case char == '.', char == '_', char == ':', char == '-':
		default:
			return false
		}
	}

	return true
}
