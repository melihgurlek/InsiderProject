package middleware

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

// Validator defines the interface for request validation.
type Validator interface {
	Validate(ctx context.Context, r *http.Request, v interface{}) error
}

// JSONValidator is a default implementation of Validator for JSON payloads.
type JSONValidator struct{}

// Validate decodes and validates a JSON request body into v.
func (jv *JSONValidator) Validate(ctx context.Context, r *http.Request, v interface{}) error {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		return ErrInvalidContentType
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, v); err != nil {
		return err
	}
	// Optionally, add struct-level validation here (e.g., required fields)
	return nil
}

// ErrInvalidContentType is returned when the request is not JSON.
var ErrInvalidContentType = &ValidationError{"invalid content type, expected application/json"}

// ValidationError represents a request validation error.
type ValidationError struct {
	Msg string
}

func (e *ValidationError) Error() string { return e.Msg }

// ValidationMiddleware returns a middleware that validates requests using the provided Validator.
func ValidationMiddleware(validator Validator, vFactory func() interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			v := vFactory()
			if err := validator.Validate(r.Context(), r, v); err != nil {
				http.Error(w, "validation error: "+err.Error(), http.StatusBadRequest)
				return
			}
			// Store validated struct in context for handler use
			ctx := context.WithValue(r.Context(), validatedBodyKey{}, v)

			// Debug: Log what we're storing
			log.Printf("DEBUG: Storing validated body of type %T in context", v)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// validatedBodyKey is the context key for the validated request body.
type validatedBodyKey struct{}

// GetValidatedBody retrieves the validated request body from context.
func GetValidatedBody[T any](ctx context.Context) (T, bool) {
	v, ok := ctx.Value(validatedBodyKey{}).(T)
	log.Printf("DEBUG: Retrieving from context, found: %v, ok: %v, type: %T", v, ok, v)
	return v, ok
}
