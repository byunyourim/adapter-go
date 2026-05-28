// Package evm 은 EVM 공통 유틸 도메인이다.
// (TS의 domain/shared/evm-address, erc20-selectors, user-op.helper 대응)
// CREATE2 주소 계산, EIP-55 checksum, ERC-20 셀렉터, UserOperation 헬퍼.
package evm

// ToChecksum 은 주소를 EIP-55 checksum 형식으로 변환한다.
//
// TODO(골격): go-ethereum common.Address 기반 구현.
func ToChecksum(address string) string {
	panic("not implemented")
}

// PredictCreate2 는 CREATE2 결정론적 주소를 계산한다(배포 전 입금 주소 예측).
//
// TODO(골격): keccak256(0xff ++ deployer ++ salt ++ keccak256(initCode))[12:]
func PredictCreate2(deployer, salt, initCodeHash string) string {
	panic("not implemented")
}
