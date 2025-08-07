package pkg

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/melihgurlek/backend-path/internal/middleware"
)

// JWTValidatorImpl implements the JWTValidator interface for validating JWT tokens.
type JWTValidatorImpl struct {
	secret string
}

// NewJWTValidator creates a new JWTValidatorImpl with the given secret key.
func NewJWTValidator(secret string) *JWTValidatorImpl {
	return &JWTValidatorImpl{secret: secret}
}

// ValidateToken parses and validates a JWT token string, returning user claims if valid.
func (j *JWTValidatorImpl) ValidateToken(tokenString string) (*middleware.UserClaims, error) {
	errWrongMethod := errors.New("unexpected signing method")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errWrongMethod
		}
		return []byte(j.secret), nil
	})
	if err != nil {
		if strings.Contains(err.Error(), errWrongMethod.Error()) {
			return nil, errWrongMethod
		}
		return nil, errors.New("invalid token")
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	// Explicitly check the algorithm
	if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
		return nil, errWrongMethod
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims type")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("user_id claim missing or invalid")
	}
	role, ok := claims["role"].(string)
	if !ok {
		return nil, errors.New("role claim missing or invalid")
	}

	jti, ok := claims["jti"].(string)
	if !ok {
		return nil, errors.New("jti claim missing or invalid")
	}

	return &middleware.UserClaims{
		UserID: userID,
		Role:   role,
		JTI:    jti,
	}, nil
}

// GenerateToken creates a new JWT token with the given user claims.
func GenerateToken(secret string, userID string, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"jti":     uuid.New().String(),
		"exp":     time.Now().Add(15 * time.Minute).Unix(), // 15 minute expiration
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
