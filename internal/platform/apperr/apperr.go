// Package apperr 어댑터 공통 에러 체계 (TS의 domain/error/* 대응)
package apperr

import (
	"errors"
	"log/slog"
	"strings"
)

// ErrorCode 상수 — TS domain/error/error-code.ts와 1:1
const (
	// Validation
	CodeValidationError      = "VALIDATION_ERROR"
	CodeMissingRequiredField = "MISSING_REQUIRED_FIELD"
	CodeInvalidString        = "INVALID_STRING"
	CodeInvalidEnumValue     = "INVALID_ENUM_VALUE"
	CodeInvalidPositiveInt   = "INVALID_POSITIVE_INT"
	CodeInvalidAddressFormat = "INVALID_ADDRESS_FORMAT"
	CodeInvalidAmount        = "INVALID_AMOUNT"
	CodeUnsupportedChain     = "UNSUPPORTED_CHAIN"

	// Not Found
	CodeNotFound        = "NOT_FOUND"
	CodeAccountNotFound = "ACCOUNT_NOT_FOUND"
	CodeTokenNotFound   = "TOKEN_NOT_FOUND"
	CodeSenderNotFound  = "SENDER_NOT_FOUND"

	// RPC
	CodeRPCConnectionTimeout = "RPC_CONNECTION_TIMEOUT"
	CodeRPCConnectionRefused = "RPC_CONNECTION_REFUSED"
	CodeRPCNotConfigured     = "RPC_NOT_CONFIGURED"
	CodeRPCCallFailed        = "RPC_CALL_FAILED"

	// Bundler
	CodeBundlerBuildFailed   = "BUNDLER_BUILD_FAILED"
	CodeBundlerSendFailed    = "BUNDLER_SEND_FAILED"
	CodeBundlerReceiptFailed = "BUNDLER_RECEIPT_FAILED"
	CodeBundlerNotConfigured = "BUNDLER_NOT_CONFIGURED"

	// Kafka
	CodeKafkaPublishFailed = "KAFKA_PUBLISH_FAILED"

	// DB
	CodeDBSaveFailed  = "DB_SAVE_FAILED"
	CodeDBQueryFailed = "DB_QUERY_FAILED"

	// Lock
	CodeLockAcquisitionFailed = "LOCK_ACQUISITION_FAILED"

	// Business
	CodeBusinessError       = "BUSINESS_ERROR"
	CodeNoAvailableNetwork  = "NO_AVAILABLE_NETWORK"
	CodeUnsupportedNetwork  = "UNSUPPORTED_NETWORK"
	CodeInsufficientBalance = "INSUFFICIENT_BALANCE"
	CodeDuplicateRequest    = "DUPLICATE_REQUEST"
	CodeAlreadyDeployed     = "ALREADY_DEPLOYED"
	CodeAddressMismatch     = "ADDRESS_MISMATCH"
	CodeTransactionFailed   = "TRANSACTION_FAILED"

	// Unknown
	CodeUnknownError = "UNKNOWN_ERROR"
)

// DefaultMessage ErrorCode별 기본 메시지 — TS DefaultMessage와 동일
var DefaultMessage = map[string]string{
	CodeValidationError:      "Validation failed",
	CodeMissingRequiredField: "Required field is missing",
	CodeInvalidString:        "Must be a non-empty string",
	CodeInvalidEnumValue:     "Value is not in the allowed list",
	CodeInvalidPositiveInt:   "Must be a positive integer",
	CodeInvalidAddressFormat: "Invalid blockchain address format",
	CodeInvalidAmount:        "Amount must be a positive number",
	CodeUnsupportedChain:     "Unsupported blockchain chain",

	CodeNotFound:        "Requested resource not found",
	CodeAccountNotFound: "Account not found",
	CodeTokenNotFound:   "Token not found",
	CodeSenderNotFound:  "Sender account not found",

	CodeRPCConnectionTimeout: "RPC connection timed out",
	CodeRPCConnectionRefused: "RPC connection refused",
	CodeRPCNotConfigured:     "RPC URL is not configured",
	CodeRPCCallFailed:        "RPC call failed",

	CodeBundlerBuildFailed:   "Failed to build UserOperation",
	CodeBundlerSendFailed:    "Failed to send UserOperation to bundler",
	CodeBundlerReceiptFailed: "Failed to retrieve UserOperation receipt",
	CodeBundlerNotConfigured: "Bundler URL is not configured",

	CodeKafkaPublishFailed: "Failed to publish message to Kafka",

	CodeDBSaveFailed:  "Failed to save data to database",
	CodeDBQueryFailed: "Failed to query data from database",

	CodeLockAcquisitionFailed: "Failed to acquire distributed lock, please retry",

	CodeBusinessError:       "Business logic error",
	CodeNoAvailableNetwork:  "No available network configured",
	CodeUnsupportedNetwork:  "Unsupported network type",
	CodeInsufficientBalance: "Insufficient balance",
	CodeDuplicateRequest:    "Duplicate request detected",
	CodeAlreadyDeployed:     "Account is already deployed",
	CodeAddressMismatch:     "Deployed address does not match expected address",
	CodeTransactionFailed:   "Transaction failed on-chain (possible causes: insufficient gas, contract revert, ERC20 transfer failure)",

	CodeUnknownError: "An unexpected error occurred",
}

// AppError 카테고리를 StatusCode/Retryable로 구분
type AppError struct {
	Code       string
	StatusCode int
	Retryable  bool
	Field      string // ValidationError.field
	Identifier string // NotFoundError.identifier
	msg        string // 메시지 override (없으면 DefaultMessage[Code])
	cause      error
}

// WithMessage 기본 메시지 대신 쓸 메시지 지정(세분 코드별 설명 등)
func (e *AppError) WithMessage(m string) *AppError {
	e.msg = m
	return e
}

// Error 메시지(override > DefaultMessage > code) + field/identifier 접미 반환
func (e *AppError) Error() string {
	msg := e.msg
	if msg == "" {
		var ok bool
		if msg, ok = DefaultMessage[e.Code]; !ok {
			msg = e.Code
		}
	}
	if e.Field != "" {
		msg += " (field: " + e.Field + ")"
	}
	if e.Identifier != "" {
		msg += " (" + e.Identifier + ")"
	}
	return msg
}

// Unwrap cause 노출로 errors.Is/As 체인 연결
func (e *AppError) Unwrap() error { return e.cause }

// LogValue slog가 AppError를 구조로 남기게 함(ELK code 필드)
func (e *AppError) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("code", e.Code),
		slog.Int("statusCode", e.StatusCode),
		slog.Bool("retryable", e.Retryable),
	}
	if e.Field != "" {
		attrs = append(attrs, slog.String("field", e.Field))
	}
	if e.Identifier != "" {
		attrs = append(attrs, slog.String("identifier", e.Identifier))
	}
	if e.cause != nil {
		attrs = append(attrs, slog.String("cause", e.cause.Error()))
	}
	return slog.GroupValue(attrs...)
}

// New 기본 AppError(500, 비재시도) 생성
func New(code string, cause error) *AppError {
	return &AppError{Code: code, StatusCode: 500, cause: cause}
}

// NewValidation 검증 에러(400) 생성
func NewValidation(code, field string, cause error) *AppError {
	if code == "" {
		code = CodeValidationError
	}
	return &AppError{Code: code, StatusCode: 400, Field: field, cause: cause}
}

// NewNotFound 미발견 에러(404) 생성
func NewNotFound(code, identifier string, cause error) *AppError {
	if code == "" {
		code = CodeNotFound
	}
	return &AppError{Code: code, StatusCode: 404, Identifier: identifier, cause: cause}
}

// NewBusiness 비즈니스 규칙 위반 에러(422) 생성
func NewBusiness(code string, cause error) *AppError {
	if code == "" {
		code = CodeBusinessError
	}
	return &AppError{Code: code, StatusCode: 422, cause: cause}
}

// NewInfra 인프라 장애 에러(502, 재시도 가능) 생성
func NewInfra(code string, cause error) *AppError {
	if code == "" {
		code = CodeUnknownError
	}
	return &AppError{Code: code, StatusCode: 502, Retryable: true, cause: cause}
}

// WrapInfra 예외를 InfraError로 래핑 — 이미 AppError면 그대로 반환
func WrapInfra(err error, code string) error {
	var ae *AppError
	if errors.As(err, &ae) {
		return err
	}
	return NewInfra(code, err)
}

// Code 에러의 분류 코드 반환(구조화 로그의 code 필드용)
func Code(err error) string {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.Code
	}
	if err == nil {
		return ""
	}
	return CodeUnknownError
}

// retryableKeywords 비-AppError 네트워크 에러 판별용 키워드
var retryableKeywords = []string{
	"econnrefused", "econnreset", "etimedout", "enetunreach",
	"eai_again", "socket hang up", "network error",
}

// IsRetryable 재시도 가능 여부 판단
// AppError면 Retryable 속성, 그 외 Error는 네트워크 키워드 매칭
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.Retryable
	}
	msg := strings.ToLower(err.Error())
	for _, k := range retryableKeywords {
		if strings.Contains(msg, k) {
			return true
		}
	}
	return false
}
