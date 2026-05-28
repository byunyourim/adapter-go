package bundler

import (
	"errors"
	"testing"

	"github.com/byunyourim/stablecoinbc-adapter/internal/platform/apperr"
)

func TestClassify_RevertIsNonRetryableBusiness(t *testing.T) {
	// 번들러 응답 cause에서 revert 코드 추출 (TS "→ N: CODE" 패턴)
	err := Classify(errors.New(`UserOp reverted → 3: AC007`))
	if apperr.Code(err) != "AC007" {
		t.Errorf("code=%q, want AC007", apperr.Code(err))
	}
	if apperr.IsRetryable(err) {
		t.Error("revert(AC007)는 비재시도 기대 → DLQ")
	}
	if err.Error() != "insufficient ERC20 balance" {
		t.Errorf("message=%q", err.Error())
	}
}

func TestClassify_ReasonAndCodePatterns(t *testing.T) {
	if got := apperr.Code(Classify(errors.New(`execution reverted reason="AC010"`))); got != "AC010" {
		t.Errorf("reason 패턴 code=%q, want AC010", got)
	}
	if got := apperr.Code(Classify(errors.New(`error code=CALL_EXCEPTION blah`))); got != "CALL_EXCEPTION" {
		t.Errorf("code 패턴 code=%q, want CALL_EXCEPTION", got)
	}
}

func TestClassify_NonceConflictIsRetryable(t *testing.T) {
	err := Classify(errors.New("AA25 invalid account nonce"))
	if apperr.Code(err) != CodeNonceConflict {
		t.Errorf("code=%q, want %q", apperr.Code(err), CodeNonceConflict)
	}
	if !apperr.IsRetryable(err) {
		t.Error("nonce 충돌은 재시도 기대")
	}
	if !IsNonceConflict(err) {
		t.Error("IsNonceConflict true 기대")
	}
}

func TestClassify_UnknownFallsBackToSendFailed(t *testing.T) {
	err := Classify(errors.New("connection reset by peer"))
	if apperr.Code(err) != apperr.CodeBundlerSendFailed {
		t.Errorf("code=%q, want %q", apperr.Code(err), apperr.CodeBundlerSendFailed)
	}
	if !apperr.IsRetryable(err) {
		t.Error("미상 전송 오류는 재시도 기대")
	}
}

func TestClassify_NilReturnsNil(t *testing.T) {
	if Classify(nil) != nil {
		t.Error("nil 입력은 nil 반환 기대")
	}
}

// 번들러 구조화 응답(tx-error.ts)을 신뢰해 분류.
func TestClassifyResponse_StructuredFields(t *testing.T) {
	cause := errors.New("bundler 422")

	// revert 코드 + 비재시도 → business
	rev := ClassifyResponse("AC007", false, cause)
	if apperr.Code(rev) != "AC007" || apperr.IsRetryable(rev) {
		t.Errorf("AC007: code=%q retryable=%v, want AC007/false", apperr.Code(rev), apperr.IsRetryable(rev))
	}
	if rev.Error() != "insufficient ERC20 balance" {
		t.Errorf("message=%q", rev.Error())
	}

	// nonce 충돌 + 재시도 → infra, IsNonceConflict
	nc := ClassifyResponse(CodeNonceConflict, true, cause)
	if !apperr.IsRetryable(nc) || !IsNonceConflict(nc) {
		t.Errorf("nonce: retryable=%v isNonce=%v, want true/true", apperr.IsRetryable(nc), IsNonceConflict(nc))
	}

	// 비-revert 코드도 친화 메시지 매핑됨(코드 문자열 그대로 아님)
	to := ClassifyResponse(CodeRPCTimeout, true, cause)
	if apperr.Code(to) != CodeRPCTimeout || !apperr.IsRetryable(to) {
		t.Errorf("RPC_TIMEOUT: code=%q retryable=%v", apperr.Code(to), apperr.IsRetryable(to))
	}
	if to.Error() == CodeRPCTimeout {
		t.Errorf("RPC_TIMEOUT 메시지가 코드 그대로임 — 매핑 누락: %q", to.Error())
	}

	// code 비면 정규식 폴백
	fb := ClassifyResponse("", true, errors.New("→ 3: AC010"))
	if apperr.Code(fb) != "AC010" {
		t.Errorf("폴백 code=%q, want AC010", apperr.Code(fb))
	}
}
