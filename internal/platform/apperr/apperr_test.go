package apperr

import (
	"errors"
	"fmt"
	"testing"
)

func TestCategoryDefaults(t *testing.T) {
	cases := []struct {
		name       string
		err        *AppError
		statusCode int
		retryable  bool
	}{
		{"validation", NewValidation(CodeMissingRequiredField, "address", nil), 400, false},
		{"notfound", NewNotFound(CodeAccountNotFound, "0xabc", nil), 404, false},
		{"business", NewBusiness(CodeAlreadyDeployed, nil), 422, false},
		{"infra", NewInfra(CodeRPCCallFailed, nil), 502, true},
		{"base", New(CodeUnknownError, nil), 500, false},
	}
	for _, c := range cases {
		if c.err.StatusCode != c.statusCode {
			t.Errorf("%s: statusCode=%d, want %d", c.name, c.err.StatusCode, c.statusCode)
		}
		if c.err.Retryable != c.retryable {
			t.Errorf("%s: retryable=%v, want %v", c.name, c.err.Retryable, c.retryable)
		}
	}
}

func TestErrorMessage(t *testing.T) {
	v := NewValidation(CodeMissingRequiredField, "address", nil)
	want := "Required field is missing (field: address)"
	if v.Error() != want {
		t.Errorf("Error()=%q, want %q", v.Error(), want)
	}

	nf := NewNotFound(CodeAccountNotFound, "0xabc", nil)
	if nf.Error() != "Account not found (0xabc)" {
		t.Errorf("Error()=%q", nf.Error())
	}
}

func TestCode(t *testing.T) {
	if got := Code(NewBusiness(CodeAlreadyDeployed, nil)); got != CodeAlreadyDeployed {
		t.Errorf("Code=%q, want %q", got, CodeAlreadyDeployed)
	}
	// %w로 래핑돼도 errors.As로 찾는다
	wrapped := fmt.Errorf("context: %w", NewInfra(CodeDBSaveFailed, nil))
	if got := Code(wrapped); got != CodeDBSaveFailed {
		t.Errorf("wrapped Code=%q, want %q", got, CodeDBSaveFailed)
	}
	if got := Code(errors.New("plain")); got != CodeUnknownError {
		t.Errorf("plain Code=%q, want %q", got, CodeUnknownError)
	}
	if got := Code(nil); got != "" {
		t.Errorf("nil Code=%q, want empty", got)
	}
}

func TestIsRetryable(t *testing.T) {
	if !IsRetryable(NewInfra(CodeRPCCallFailed, nil)) {
		t.Error("infra는 retryable 기대")
	}
	if IsRetryable(NewBusiness(CodeAlreadyDeployed, nil)) {
		t.Error("business는 non-retryable 기대")
	}
	// AppError 아닌 네트워크 에러 키워드 매칭
	if !IsRetryable(errors.New("dial tcp: ECONNREFUSED")) {
		t.Error("ECONNREFUSED는 retryable 기대")
	}
	if IsRetryable(errors.New("some logic error")) {
		t.Error("일반 에러는 non-retryable 기대")
	}
}

func TestWrapInfra(t *testing.T) {
	// 이미 AppError면 그대로 통과
	orig := NewBusiness(CodeInsufficientBalance, nil)
	if got := WrapInfra(orig, CodeDBQueryFailed); got != error(orig) {
		t.Error("AppError는 그대로 반환 기대")
	}
	// 일반 에러는 InfraError로 래핑 + cause 보존
	plain := errors.New("boom")
	wrapped := WrapInfra(plain, CodeDBQueryFailed)
	if Code(wrapped) != CodeDBQueryFailed {
		t.Errorf("code=%q, want %q", Code(wrapped), CodeDBQueryFailed)
	}
	if !errors.Is(wrapped, plain) {
		t.Error("cause 보존(errors.Is) 기대")
	}
	if !IsRetryable(wrapped) {
		t.Error("InfraError는 retryable 기대")
	}
}
