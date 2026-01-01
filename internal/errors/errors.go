package errors

import (
	"encoding/json"
	Errors "errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

type AppError struct {
	Code    ErrorCode `json:"error_code"`
	Message string    `json:"message"`
	Cause   error     `json:"cause"`
}

type ServerError struct {
	AppError
	StatusCode int `json:"status_code"`
}

func NewAppError(code ErrorCode, message string, cause error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// errorJSON represents the JSON structure of an error
type errorJSON struct {
	ErrorCode string      `json:"error_code"`
	Message   string      `json:"message"`
	Cause     interface{} `json:"cause,omitempty"`
}

// serverErrorJSON represents the JSON structure of a ServerError
type serverErrorJSON struct {
	ErrorCode  string      `json:"error_code"`
	Message    string      `json:"message"`
	StatusCode int         `json:"status_code"`
	Cause      interface{} `json:"cause,omitempty"`
}

// toErrorJSON converts an error to a JSON-serializable structure
func toErrorJSON(err error) interface{} {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if Errors.As(err, &appErr) {
		// Double-check that appErr is not nil
		if appErr == nil {
			return err
		}

		result := errorJSON{
			ErrorCode: string(appErr.Code),
			Message:   appErr.Message,
		}
		if appErr.Cause != nil {
			result.Cause = toErrorJSON(appErr.Cause)
		}
		return result
	}

	// For non-AppError errors, safely get the error message
	// Use fmt.Sprintf with %v to safely handle any error type
	return fmt.Sprintf("%v", err)
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}

	jsonStruct := errorJSON{
		ErrorCode: string(e.Code),
		Message:   e.Message,
	}
	if e.Cause != nil {
		jsonStruct.Cause = toErrorJSON(e.Cause)
	}

	jsonBytes, err := json.Marshal(jsonStruct)
	if err != nil {
		// Fallback to simple format if JSON marshaling fails
		return fmt.Sprintf(`{"error_code":"%s","message":"%s"}`, e.Code, e.Message)
	}
	return string(jsonBytes)
}

func (e *AppError) String() string {
	return e.Error()
}

// MarshalLogObject implements zapcore.ObjectMarshaler for proper JSON logging
func (e *AppError) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if e == nil {
		return nil
	}
	enc.AddString("error_code", string(e.Code))
	enc.AddString("message", e.Message)
	if e.Cause != nil {
		// Recursively marshal the cause if it's also an AppError
		var appErr *AppError
		if Errors.As(e.Cause, &appErr) {
			return enc.AddObject("cause", appErr)
		}
		// For non-AppError causes, add as string
		enc.AddString("cause", e.Cause.Error())
	}
	return nil
}

func Wrap(err error, code ErrorCode, message string) *AppError {
	if err == nil {
		return nil
	}
	return NewAppError(code, message, err)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// inherits the error and returns a new AppError with the same cause and message
func Inherit(err error, code ErrorCode) *AppError {
	if err == nil {
		return nil
	}
	var appErr *AppError
	if Errors.As(err, &appErr) {
		return NewAppError(code, appErr.Message, appErr.Cause)
	}
	return NewAppError(code, err.Error(), err)
}
func (e *AppError) Is(target error) bool {
	if t, ok := target.(*AppError); ok {
		return e.Code == t.Code
	}
	return Errors.Is(e.Cause, target)
}

func (e *AppError) GetCategory() string {
	// Separate the code from "-"
	parts := strings.Split(string(e.Code), "-")
	return parts[0]
}

func (e *AppError) As(target interface{}) bool {
	if t, ok := target.(**AppError); ok {
		*t = e
		return true
	}
	return Errors.As(e.Cause, target)
}

// HasCode checks if error has a specific error code (traverses entire chain)
func HasCode(err error, code ErrorCode) bool {
	// Traverse the error chain (via Cause/Unwrap)
	for err != nil {
		var appErr *AppError
		// Use standard library errors.As() - it traverses the chain automatically
		if Errors.As(err, &appErr) {
			if appErr.Code == code {
				return true
			}
		}
		// Move to next error in chain
		err = Errors.Unwrap(err)
	}
	return false
}

func HasAppError(err error) bool {
	var appErr *AppError
	return Errors.As(err, &appErr)
}

// GetRootAppError finds the root AppError in an error chain
// Works with any error type, traverses via Unwrap()
func GetRootAppError(err error) *AppError {
	// Find first AppError in chain
	var root *AppError
	if !Errors.As(err, &root) {
		return nil
	}

	// Traverse to deepest AppError
	for {
		unwrapped := Errors.Unwrap(root)
		// Check if unwrapped is nil before calling Errors.As
		if unwrapped == nil {
			break // End of error chain
		}
		var next *AppError
		if !Errors.As(unwrapped, &next) {
			break // End of AppError chain
		}
		root = next
	}

	return root
}

// Method version for convenience
func (e *AppError) GetRoot() *AppError {
	if e == nil {
		return nil // Handle nil receiver
	}
	return GetRootAppError(e)
}

// GetAllCodes can also use Unwrap()
func GetAllCodes(err error) []ErrorCode {
	var codes []ErrorCode

	// Traverse entire chain
	current := err
	for current != nil {
		var appErr *AppError
		if Errors.As(current, &appErr) {
			codes = append(codes, appErr.Code)
		}
		current = Errors.Unwrap(current)
	}
	return codes
}

// Method version
func (e *AppError) GetAllCodes() []ErrorCode {
	return GetAllCodes(e)
}

func (e *AppError) ToServerError(statusCode int) *ServerError {
	if e == nil {
		return nil
	}
	return NewServerError(e.Code, e.Message, e.Cause, statusCode)
}

func NewServerError(code ErrorCode, message string, cause error, statusCode int) *ServerError {
	return &ServerError{
		AppError:   *NewAppError(code, message, cause),
		StatusCode: statusCode,
	}
}

func (e *ServerError) Error() string {
	if e == nil {
		return ""
	}

	jsonStruct := serverErrorJSON{
		ErrorCode:  string(e.Code),
		Message:    e.Message,
		StatusCode: e.StatusCode,
	}
	if e.Cause != nil {
		jsonStruct.Cause = toErrorJSON(e.Cause)
	}

	jsonBytes, err := json.Marshal(jsonStruct)
	if err != nil {
		// Fallback to simple format if JSON marshaling fails
		return fmt.Sprintf(`{"error_code":"%s","message":"%s","status_code":%d}`, e.Code, e.Message, e.StatusCode)
	}
	return string(jsonBytes)
}

func (e *ServerError) String() string {
	return e.Error()
}

// MarshalLogObject implements zapcore.ObjectMarshaler for proper JSON logging
func (e *ServerError) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if e == nil {
		return nil
	}
	enc.AddString("error_code", string(e.Code))
	enc.AddString("message", e.Message)
	enc.AddInt("status_code", e.StatusCode)
	if e.Cause != nil {
		// Recursively marshal the cause if it's also an AppError
		var appErr *AppError
		if Errors.As(e.Cause, &appErr) {
			return enc.AddObject("cause", appErr)
		}
		// For non-AppError causes, add as string
		enc.AddString("cause", e.Cause.Error())
	}
	return nil
}

func (e *ServerError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.AppError.Unwrap()
}

func (e *ServerError) Is(target error) bool {
	if e == nil {
		return false
	}
	return e.AppError.Is(target)
}

func (e *ServerError) As(target interface{}) bool {
	if e == nil {
		return false
	}
	return e.AppError.As(target)
}

func WrapServer(err error, code ErrorCode, message string, statusCode int) *ServerError {
	if err == nil {
		return nil
	}
	return NewServerError(code, message, err, statusCode)
}

func HandleDBReadError(err error) *ServerError {
	if err == nil {
		return nil
	}
	if Errors.Is(err, gorm.ErrRecordNotFound) {
		return NewServerError(DBRecordNotFound, "Record not found", err, fiber.StatusNotFound)
	}
	return NewServerError(DBQueryFailed, "Query failed", err, fiber.StatusInternalServerError)
}

func HandleDBWriteError(err error) *AppError {
	if err == nil {
		return nil
	}
	if Errors.Is(err, gorm.ErrRecordNotFound) {
		return NewAppError(DBRecordNotFound, "Record not found", err)
	}
	return NewAppError(DBQueryFailed, "Query failed", err)
}

func HandleDBCreateError(err error) *AppError {
	if err == nil {
		return nil
	}
	if Errors.Is(err, gorm.ErrInvalidTransaction) {
		return NewAppError(DBTransactionFailed, "Transaction failed", err)
	}
	if Errors.Is(err, gorm.ErrDuplicatedKey) {
		return NewAppError(DBConstraintViolation, "Constraint violation", err)
	}
	return NewAppError(DBQueryFailed, "Query failed", err)
}

// Helper function to parse ServerError from JSON response
func ParseServerError(body []byte, statusCode int) *ServerError {
	// Use a temporary struct that matches the JSON structure
	var temp struct {
		ErrorCode  string          `json:"error_code"`
		Message    string          `json:"message"`
		StatusCode int             `json:"status_code"`
		Cause      json.RawMessage `json:"cause,omitempty"`
	}

	if err := json.Unmarshal(body, &temp); err != nil {
		// If unmarshaling fails, create a ServerError with the raw body as message
		return NewServerError(
			ClientResponseInvalid,
			fmt.Sprintf("Failed to parse error response: %s", string(body)),
			fmt.Errorf("json unmarshal failed: %w", err),
			statusCode,
		)
	}

	// Parse the cause if it exists
	var cause error
	if len(temp.Cause) > 0 {
		// Try to unmarshal cause as an AppError first
		var causeAppErr struct {
			ErrorCode string          `json:"error_code"`
			Message   string          `json:"message"`
			Cause     json.RawMessage `json:"cause,omitempty"`
		}
		if err := json.Unmarshal(temp.Cause, &causeAppErr); err == nil {
			// It's an AppError structure
			var nestedCause error
			if len(causeAppErr.Cause) > 0 {
				nestedCause = fmt.Errorf("%s", string(causeAppErr.Cause))
			}
			cause = NewAppError(
				ErrorCode(causeAppErr.ErrorCode),
				causeAppErr.Message,
				nestedCause,
			)
		} else {
			// It's a plain string or other type
			var causeStr string
			if err := json.Unmarshal(temp.Cause, &causeStr); err == nil {
				cause = fmt.Errorf(causeStr)
			} else {
				cause = fmt.Errorf("%s", string(temp.Cause))
			}
		}
	}

	return NewServerError(
		ErrorCode(temp.ErrorCode),
		temp.Message,
		cause,
		temp.StatusCode,
	)
}
