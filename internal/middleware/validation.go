package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
)

// Validatable is an interface for structs that can be validated.
type Validatable interface {
	Validate() error
}

type Validator interface {
	Validate(ctx context.Context, r *http.Request, v interface{}) error
}

type JSONValidator struct{}

func (jv *JSONValidator) Validate(ctx context.Context, r *http.Request, v interface{}) error {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		return ErrInvalidContentType
	}

	// Read the body and then replace it so it can be read again
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	r.Body.Close()
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := json.Unmarshal(bodyBytes, v); err != nil {
		return &ValidationError{Msg: "invalid JSON format"}
	}

	// Check if the decoded struct implements the Validatable interface
	if validatable, ok := v.(Validatable); ok {
		if err := validatable.Validate(); err != nil {
			return err
		}
	} else {
		// Also check if the pointer to the struct implements it
		ptr := reflect.New(reflect.TypeOf(v).Elem())
		ptr.Elem().Set(reflect.ValueOf(v).Elem())
		if validatable, ok := ptr.Interface().(Validatable); ok {
			if err := validatable.Validate(); err != nil {
				return err
			}
		}
	}

	return nil
}

var ErrInvalidContentType = &ValidationError{Msg: "invalid content type, expected application/json"}

type ValidationError struct {
	Msg string
}

func (e *ValidationError) Error() string { return e.Msg }

func ValidationMiddleware(validator Validator, vFactory func() interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			v := vFactory()
			if err := validator.Validate(r.Context(), r, &v); err != nil {
				// Return a 400 Bad Request for any validation error
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			ctx := context.WithValue(r.Context(), validatedBodyKey{}, v)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type validatedBodyKey struct{}

func GetValidatedBody[T any](ctx context.Context) (T, bool) {
	v, ok := ctx.Value(validatedBodyKey{}).(T)
	return v, ok
}
