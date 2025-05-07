package constancy

import (
	"database/sql"
	"github.com/uptrace/bun"
)

// PGTxWrapper pg.Tx
//
//go:generate mockgen -source=storage_wrapper.go -destination=../test/mock/constancy/stub_storage_wrapper.go -package=mock_constancy -imports pg=github.com/uptrace/bun
type PGTxWrapper interface {
	Rollback(*bun.Tx) error
	Commit(*bun.Tx) error
}

type PGTxWrapperImpl struct {
}

func (db *PGTxWrapperImpl) Rollback(tx *bun.Tx) error {
	return tx.Rollback()
}

func (db *PGTxWrapperImpl) Commit(tx *bun.Tx) error {
	return tx.Commit()
}

// PGConnWrapper pg.Conn
type PGConnWrapper interface {
	Model(conn *bun.SelectQuery, model interface{}) *bun.SelectQuery
	Begin(conn *bun.Conn) (*bun.Tx, error)
	Close(conn *bun.Conn) error
}

type PGConnWrapperImpl struct {
}

func (db *PGConnWrapperImpl) Model(conn *bun.SelectQuery, model interface{}) *bun.SelectQuery {
	return conn.Model(model)
}

func (db *PGConnWrapperImpl) Begin(conn *bun.Conn) (*bun.Tx, error) {
	tx, err := conn.BeginTx(ctx, &sql.TxOptions{})
	return &tx, err
}

func (db *PGConnWrapperImpl) Close(conn *bun.Conn) error {
	return conn.Close()
}

// PGQueryWrapper pg.Query
type PGQueryWrapper interface {
	Select(conn *bun.Conn) *bun.SelectQuery
	Where(query *bun.SelectQuery, condition string, params interface{}) *bun.SelectQuery
	Scan(query *bun.SelectQuery) error
}

type PGQueryWrapperImpl struct {
}

func (q *PGQueryWrapperImpl) Select(conn *bun.Conn) *bun.SelectQuery {
	return conn.NewSelect()
}

func (q *PGQueryWrapperImpl) Where(query *bun.SelectQuery, condition string, params interface{}) *bun.SelectQuery {
	return query.Where(condition, params)
}

func (q *PGQueryWrapperImpl) Scan(query *bun.SelectQuery) error {
	return query.Scan(ctx)
}
