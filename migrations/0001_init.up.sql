-- 어댑터 시작 스키마(골격). 실제 테이블은 기존 TS 어댑터 스키마 이식하며 확장한다.

-- 입금 지갑 계좌. CREATE2 주소는 배포 전에도 예측 가능하므로 deployed로 상태 구분.
CREATE TABLE account (
    chain_id        BIGINT       NOT NULL,
    address         TEXT         NOT NULL,             -- EIP-55 checksum
    salt            TEXT         NOT NULL,
    deployed        BOOLEAN      NOT NULL DEFAULT false,
    deploy_tx_hash  TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    PRIMARY KEY (chain_id, address)
);
-- 주소 조회는 소문자 비교(listener AccountStore와 동일 규칙). chain_id scope 필수.
CREATE INDEX idx_account_addr ON account (chain_id, lower(address));

-- 체인/토큰 설정(config-register로 등록).
CREATE TABLE chain_config (
    chain_id    BIGINT       PRIMARY KEY,
    name        TEXT         NOT NULL,
    is_kcp      BOOLEAN      NOT NULL DEFAULT false,
    rpc_url     TEXT         NOT NULL,
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE token_config (
    chain_id         BIGINT  NOT NULL REFERENCES chain_config(chain_id),
    symbol           TEXT    NOT NULL,
    contract_address TEXT    NOT NULL DEFAULT '',      -- 네이티브는 빈 문자열
    decimals         INT     NOT NULL,
    PRIMARY KEY (chain_id, contract_address)
);
