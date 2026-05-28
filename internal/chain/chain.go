// Package chain 여러 기능 슬라이스가 공유하는 체인/토큰 도메인
// (TS의 domain/model/chain + domain/shared/native-symbols, token-decimals 대응)
// 외부 IO 없는 순수 도메인 — 기능 슬라이스에서 import 가능, 역방향 금지
package chain

// Chain 체인 메타
type Chain struct {
	ChainID int64
	Name    string
	IsKCP   bool
}

// Token 토큰 메타
type Token struct {
	Symbol          string
	ContractAddress string // 네이티브는 빈 문자열
	Decimals        int
}

// NormalizeSymbol KCP 체인의 wrapped 토큰 심볼을 네이티브 심볼로 변환
// (WETH→ETH 등). KCP가 아니면 원본 유지
//
// TODO(골격): 기존 listener event-parser 규칙 이식
func NormalizeSymbol(isKCP bool, raw string) string {
	panic("not implemented")
}
