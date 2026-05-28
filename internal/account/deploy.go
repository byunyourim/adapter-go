package account

import (
	"context"
	"fmt"

	"github.com/byunyourim/stablecoinbc-adapter/internal/platform/apperr"
)

// HandleDeploy 지갑 배포 명령 처리. Kafka 컨슈머에서 호출되므로
// 요청 경로(입금 감지)를 막지 않음 — "배포 중 다른 요청 지연" 해결의 핵심.
//
// 직렬화 격리: 같은 디플로이어(체인 단위)의 배포끼리만 락으로 순서를 보장하고,
// 다른 체인 배포는 병렬 진행. nonce는 Deployer 구현이 원자적으로 관리.
//
// 멱등성(at-least-once 대비): Kafka 재배달·중복 명령에도 한 주소는 한 번만 배포.
// 이미 배포된 주소면 재배포 없이 성공 결과만 재발행.
func (s *Service) HandleDeploy(ctx context.Context, cmd DeployCommand) error {
	if done, err := s.alreadyDeployed(ctx, cmd); err != nil {
		return err
	} else if done {
		return nil // 락 밖 빠른 경로 — 이미 처리됨
	}

	// 체인 단위 락으로 디플로이어 nonce 직렬화를 가둔다.
	return s.locker.WithLock(ctx, deployLockKey(cmd.ChainID), func() error {
		// 락 안에서 재확인 — 동시 중복 명령 방지.
		if done, err := s.alreadyDeployed(ctx, cmd); err != nil {
			return err
		} else if done {
			return nil
		}

		// 예측 주소를 먼저 영속(미배포 상태). 이미 있으면 Save가 무시(멱등).
		if err := s.repo.Save(ctx, Account{ChainID: cmd.ChainID, Address: cmd.Address, Salt: cmd.Salt}); err != nil {
			return apperr.NewInfra(apperr.CodeDBSaveFailed,
				fmt.Errorf("chain=%d addr=%s: %w", cmd.ChainID, cmd.Address, err))
		}

		txHash, err := s.deploy.Deploy(ctx, cmd.ChainID, cmd.Salt, cmd.Address)
		if err != nil {
			// WrapInfra: Deployer가 AppError(예: revert=비재시도)면 그대로, 그 외 네트워크
			// 오류면 TRANSACTION_FAILED 인프라 에러(재시도)로. 재시도/DLQ는 consumer가 판단.
			return apperr.WrapInfra(
				fmt.Errorf("chain=%d addr=%s: %w", cmd.ChainID, cmd.Address, err),
				apperr.CodeTransactionFailed)
		}

		if err := s.repo.MarkDeployed(ctx, cmd.ChainID, cmd.Address, txHash); err != nil {
			// 온체인 배포는 성공했으나 상태 갱신 실패. 재처리 시 alreadyDeployed가 못 잡으므로
			// 에러를 올려 재시도(MarkDeployed·Deploy 모두 멱등이어야 함).
			return apperr.NewInfra(apperr.CodeDBSaveFailed,
				fmt.Errorf("mark deployed chain=%d addr=%s tx=%s: %w", cmd.ChainID, cmd.Address, txHash, err))
		}

		s.log.Info("wallet deployed", "chain", cmd.ChainID, "address", cmd.Address, "tx", txHash)
		return s.publishSuccess(ctx, cmd, txHash)
	})
}

// alreadyDeployed 이미 배포된 계좌면 성공 결과 재발행 후 true 반환
func (s *Service) alreadyDeployed(ctx context.Context, cmd DeployCommand) (bool, error) {
	existing, ok, err := s.repo.Get(ctx, cmd.ChainID, cmd.Address)
	if err != nil {
		return false, apperr.NewInfra(apperr.CodeDBQueryFailed,
			fmt.Errorf("chain=%d addr=%s: %w", cmd.ChainID, cmd.Address, err))
	}
	if !ok || !existing.Deployed {
		return false, nil
	}
	if err := s.publishSuccess(ctx, cmd, existing.DeployTxHash); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) publishSuccess(ctx context.Context, cmd DeployCommand, txHash string) error {
	r := DeployResult{ChainID: cmd.ChainID, Address: cmd.Address, TxHash: txHash, Success: true}
	if err := s.results.PublishDeployResult(ctx, r); err != nil {
		return apperr.NewInfra(apperr.CodeKafkaPublishFailed,
			fmt.Errorf("chain=%d addr=%s: %w", cmd.ChainID, cmd.Address, err))
	}
	return nil
}

// deployLockKey 체인(=디플로이어) 단위 분산락 키
func deployLockKey(chainID int64) string {
	return fmt.Sprintf("deploy:chain:%d", chainID)
}
