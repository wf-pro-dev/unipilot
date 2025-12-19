package errors

import (
	"encoding/json"
	Errors "errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
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

	jsonBytes, err := json.MarshalIndent(jsonStruct, "", "   ")
	if err != nil {
		// Fallback to simple format if JSON marshaling fails
		return fmt.Sprintf("error_code: %s, message: %s, cause: %v", e.Code, e.Message, e.Cause)
	}
	return "\n" + string(jsonBytes)
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
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
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
