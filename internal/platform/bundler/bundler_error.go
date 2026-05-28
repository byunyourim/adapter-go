package bundler

// 번들러 에러 세분화. (TS adapter/out/bundler/bundler-error.ts 이식 + 코드 승격)
//
// TS에서는 온체인 revert 사유(AC*/EP*/PM*/ethers)를 로그 문자열로만 남기고
// 에러 자체는 BUNDLER_SEND_FAILED(retryable)로 뭉뚱그렸다. 여기서는 그 사유를
// 첫급 에러 코드로 승격하고 재시도 가능 여부까지 분류한다:
//   - revert 사유(잔액 부족·서명 실패·미배포 등) → 비재시도(business) → DLQ
//   - nonce 충돌 → 재시도(infra)
//   - 전송/타임아웃/미상 → 재시도(infra, BUNDLER_SEND_FAILED)

import (
	"errors"
	"regexp"
	"strings"

	"github.com/byunyourim/stablecoinbc-adapter/internal/platform/apperr"
)

// 번들러(tx-error.ts/txerror.go)가 내보내는 비-revert 코드. 번들러와 문자열 일치 필수.
const (
	CodeNonceConflict = "BUNDLER_NONCE_CONFLICT" // nonce 경합 (재시도)
	CodeSendFailed    = "BUNDLER_SEND_FAILED"    // 미상 전송 실패 (재시도)
	CodeRPCTimeout    = "RPC_TIMEOUT"            // 재시도
	CodeRPCNetwork    = "RPC_NETWORK_ERROR"      // 재시도
	CodeRPCServer     = "RPC_SERVER_ERROR"       // 재시도
)

// codeInfo 번들러 에러 코드의 설명과 재시도 가능 여부
type codeInfo struct {
	message   string
	retryable bool
}

// bundlerCodes 번들러가 내보내는 모든 코드 → 분류.
// revert(AC*/EP*/PM*/INSUFFICIENT_FUNDS/CALL_EXCEPTION)는 비재시도(business) → DLQ,
// nonce/RPC/전송 오류는 재시도(infra). 번들러 txerror와 동일 집합이어야 한다.
var bundlerCodes = map[string]codeInfo{
	// Account Contract revert (비재시도)
	"AC001": {"caller is not EntryPoint", false},
	"AC004": {"target call failed", false},
	"AC005": {"native transfer failed", false},
	"AC006": {"beneficiary not set", false},
	"AC007": {"insufficient ERC20 balance", false},
	"AC008": {"ERC20 transfer failed", false},
	"AC009": {"forwarding disabled", false},
	"AC010": {"signature verification failed", false},
	// EntryPoint revert
	"EP001": {"insufficient deposit", false},
	"EP003": {"insufficient gas balance", false},
	"EP011": {"account not deployed", false},
	// Paymaster revert
	"PM006": {"insufficient token balance", false},
	"PM007": {"insufficient token allowance", false},
	// ethers/go-ethereum
	"INSUFFICIENT_FUNDS": {"insufficient funds for gas", false},
	"CALL_EXCEPTION":     {"contract call exception", false},
	// 비-revert (재시도) — 번들러가 정밀 분류해 내보내는 코드
	CodeNonceConflict: {"bundler nonce conflict, retry", true},
	CodeRPCTimeout:    {"bundler RPC timeout", true},
	CodeRPCNetwork:    {"bundler RPC network error", true},
	CodeRPCServer:     {"bundler RPC server error", true},
	CodeSendFailed:    {"bundler send failed", true},
}

var (
	reCause  = regexp.MustCompile(`→ \d+: ([A-Z][A-Z0-9_]+)`) // "... → 3: AC007"
	reReason = regexp.MustCompile(`reason="([^"]+)"`)
	reCode   = regexp.MustCompile(`code=([A-Z][A-Z_]+)`)
)

// parseRevertCode 번들러 에러 문자열에서 알려진 revert 코드 추출
// (TS parseBundlerErrorCode + formatBundlerError의 reason/code 패턴)
func parseRevertCode(s string) string {
	for _, re := range []*regexp.Regexp{reCause, reReason, reCode} {
		if m := re.FindStringSubmatch(s); m != nil {
			if _, ok := bundlerCodes[m[1]]; ok {
				return m[1]
			}
		}
	}
	return ""
}

// Classify 번들러 작업 실패 원인을 세분화된 AppError로 분류
func Classify(cause error) *apperr.AppError {
	if cause == nil {
		return nil
	}
	msg := cause.Error()

	if code := parseRevertCode(msg); code != "" {
		info := bundlerCodes[code]
		// revert는 비재시도(business). info.retryable이 true인 항목이 생기면 NewInfra로 분기.
		if info.retryable {
			return apperr.NewInfra(code, cause).WithMessage(info.message)
		}
		return apperr.NewBusiness(code, cause).WithMessage(info.message)
	}

	if strings.Contains(strings.ToLower(msg), "nonce") {
		return apperr.NewInfra(CodeNonceConflict, cause).WithMessage("bundler nonce conflict, retry")
	}

	return apperr.NewInfra(apperr.CodeBundlerSendFailed, cause)
}

// ClassifyResponse 번들러의 구조화 에러 응답(tx-error.ts)을 AppError로 변환
// 번들러가 ethers 필드로 이미 정밀 분류한 code/retryable 신뢰(정규식보다 우선)
// code가 비어 있으면(구버전 응답 등) cause 문자열 기반 Classify로 폴백
//
//	body: { error, code, category, retryable, txHash } (HTTP) 또는
//	       JSON-RPC error.data: { code, category, retryable }
func ClassifyResponse(code string, retryable bool, cause error) *apperr.AppError {
	if code == "" {
		return Classify(cause)
	}
	msg := code
	if info, ok := bundlerCodes[code]; ok {
		msg = info.message
	}
	if retryable {
		return apperr.NewInfra(code, cause).WithMessage(msg)
	}
	return apperr.NewBusiness(code, cause).WithMessage(msg)
}

// IsNonceConflict nonce 경합 에러인지 판단 (TS isNonceConflictError 대응)
func IsNonceConflict(err error) bool {
	var ae *apperr.AppError
	if errors.As(err, &ae) {
		return ae.Code == CodeNonceConflict
	}
	return false
}
