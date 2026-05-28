// Package schema 는 Ent 엔티티 스키마를 코드로 정의한다.
// `make generate`(go generate ./...)가 이 정의로 ent/ 클라이언트·쿼리 코드를 생성한다.
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Account 는 입금 지갑 계좌 엔티티. (migrations/0001_init account 테이블 대응)
type Account struct {
	ent.Schema
}

// Fields 는 account 컬럼.
func (Account) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("chain_id"),
		field.String("address"), // EIP-55 checksum
		field.String("salt"),    // CREATE2 salt
		field.Bool("deployed").Default(false),
		field.String("deploy_tx_hash").Optional(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

// Indexes 는 (chain_id, address) 유니크 — 중복 계좌 방어.
func (Account) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("chain_id", "address").Unique(),
	}
}
