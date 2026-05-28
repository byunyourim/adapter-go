package account

import (
	"context"
	"fmt"
)

// HandleDeploy 는 지갑 배포 명령을 처리한다. Kafka 컨슈머에서 호출되므로
// 요청 경로(입금 감지)를 막지 않는다 — "배포 중 다른 요청 지연" 해결의 핵심.
//
// 직렬화 격리: 같은 디플로이어(체인 단위)의 배포끼리만 락으로 순서를 보장하고,
// 다른 체인 배포는 병렬 진행한다. nonce는 Deployer 구현이 원자적으로 관리.
//
// 멱등성(at-least-once 대비): Kafka 재배달·중복 명령에도 한 주소는 한 번만 배포한다.
// 이미 배포된 주소면 재배포 없이 성공 결과만 재발행한다.
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
			return fmt.Errorf("account 저장 실패 chain=%d addr=%s: %w", cmd.ChainID, cmd.Address, err)
		}

		txHash, err := s.deploy.Deploy(ctx, cmd.ChainID, cmd.Salt, cmd.Address)
		if err != nil {
			// 에러를 그대로 올린다 — 재시도/DLQ 판단은 Kafka consumer 레이어가 한다.
			return fmt.Errorf("배포 트랜잭션 실패 chain=%d addr=%s: %w", cmd.ChainID, cmd.Address, err)
		}

		if err := s.repo.MarkDeployed(ctx, cmd.ChainID, cmd.Address, txHash); err != nil {
			// 온체인 배포는 성공했으나 상태 갱신 실패. 재처리 시 alreadyDeployed가 못 잡으므로
			// 결과는 발행하되 에러를 올려 재시도하게 한다(MarkDeployed는 멱등이어야 함).
			return fmt.Errorf("배포 상태 갱신 실패 chain=%d addr=%s tx=%s: %w", cmd.ChainID, cmd.Address, txHash, err)
		}

		s.log.Info("wallet deployed", "chain", cmd.ChainID, "address", cmd.Address, "tx", txHash)
		return s.publishSuccess(ctx, cmd, txHash)
	})
}

// alreadyDeployed 는 계좌가 이미 배포됐으면 성공 결과를 재발행하고 true를 반환한다.
func (s *Service) alreadyDeployed(ctx context.Context, cmd DeployCommand) (bool, error) {
	existing, ok, err := s.repo.Get(ctx, cmd.ChainID, cmd.Address)
	if err != nil {
		return false, fmt.Errorf("account 조회 실패 chain=%d addr=%s: %w", cmd.ChainID, cmd.Address, err)
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
		return fmt.Errorf("배포 결과 발행 실패 chain=%d addr=%s: %w", cmd.ChainID, cmd.Address, err)
	}
	return nil
}

// deployLockKey 는 체인(=디플로이어) 단위 분산락 키.
func deployLockKey(chainID int64) string {
	return fmt.Sprintf("deploy:chain:%d", chainID)
}
