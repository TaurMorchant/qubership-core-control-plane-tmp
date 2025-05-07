package constancy

import (
	"github.com/uptrace/bun"
)

type PgBatchStorage struct {
	conn *bun.Conn
}

func (s *PgBatchStorage) Insert(entities []interface{}) error {
	if len(entities) > 0 {
		for _, entity := range entities {
			_, err := s.conn.NewInsert().Model(entity).Exec(ctx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *PgBatchStorage) Update(entities []interface{}) error {
	if len(entities) > 0 {
		// TODO This is WA. Postgres multiple update doesn't work with unique columns. Maybe it's bug of PG or restriction
		//_, err := s.conn.Model(entities...).WherePK().Update()
		for _, entity := range entities {
			_, err := s.conn.NewUpdate().Model(entity).WherePK().Exec(ctx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *PgBatchStorage) Delete(entities []interface{}) error {
	if len(entities) > 0 {
		for _, entity := range entities {
			_, err := s.conn.NewDelete().Model(entity).WherePK().Exec(ctx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
