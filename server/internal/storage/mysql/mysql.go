package mysql

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.chensl.me/redix/server/internal/storage"
)

type Config struct {
	Username string
	Password string
	Host     string
	Port     int
	Database string
}

type mysqlStorage struct {
	db *sqlx.DB
}

type entry struct {
	ID       int64        `db:"_id"`
	Key      []byte       `db:"_key"`
	Value    []byte       `db:"_value"`
	ExpireAt sql.NullTime `db:"_expire_at"`
}

func NewStorage(cfg *Config) (storage.Interface, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := autoMigrate(db.DB); err != nil {
		return nil, err
	}

	return &mysqlStorage{
		db: db,
	}, nil
}

func (s *mysqlStorage) Close() error {
	return s.db.Close()
}

func (s *mysqlStorage) Set(key []byte, value []byte, opts storage.SetOptions) error {
	if opts.NX && opts.XX {
		return storage.ErrInvalidOpts
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = getEntryWithTx(tx, key)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	exist := err == nil

	if !exist && opts.XX {
		return storage.ErrNotExist
	} else if exist && opts.NX {
		return storage.ErrExist
	}

	e := entry{
		Key:   key,
		Value: value,
	}
	if opts.TTL > 0 {
		e.ExpireAt.Valid = true
		e.ExpireAt.Time = time.Now().Add(opts.TTL)
	}

	if exist {
		if err := updateValueWithTx(tx, &e); err != nil {
			return err
		}
		return tx.Commit()
	}

	if err := insertEntryWithTx(tx, &e); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *mysqlStorage) Get(key []byte) ([]byte, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	e, err := getEntryWithTx(tx, key)
	if err == sql.ErrNoRows {
		return nil, storage.ErrNotExist
	}
	if err != nil {
		return nil, err
	}

	if e.ExpireAt.Valid && time.Now().After(e.ExpireAt.Time) {
		if err := deleteEntryWithTx(tx, e.Key); err != nil {
			return nil, err
		}
	}

	_ = tx.Commit()
	return e.Value, nil
}

func (s *mysqlStorage) Add(key []byte, delta int) (int, error) {
	panic("not implemented") // TODO: Implement
}

func (s *mysqlStorage) Keys(pattern string) ([][]byte, error) {
	panic("not implemented") // TODO: Implement
}

func (s *mysqlStorage) Del(keys ...[]byte) (int, error) {
	query, args, err := sqlx.In(`
		DELETE FROM
			entries
		WHERE
			_key IN (?)
	`, keys)
	if err != nil {
		return 0, err
	}

	query = s.db.Rebind(query)
	res, err := s.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	count, _ := res.RowsAffected()
	return int(count), nil
}

func (s *mysqlStorage) Expire(key []byte, dur time.Duration) error {
	panic("not implemented") // TODO: Implement
}

func (s *mysqlStorage) TTL(key []byte) (int64, error) {
	panic("not implemented") // TODO: Implement
}

func (s *mysqlStorage) DropAll() error {
	panic("not implemented") // TODO: Implement
}

func getEntryWithTx(tx *sqlx.Tx, key []byte) (*entry, error) {
	var e entry
	if err := tx.Get(&e, `
		SELECT
			*
		FROM
			entries
		WHERE
			_key = ?
	`, key); err != nil {
		return nil, err
	}
	return &e, nil
}

func updateValueWithTx(tx *sqlx.Tx, e *entry) error {
	_, err := tx.Exec(`
		UPDATE
			entries
		SET
			_value = ?, _expire_at = ?
		WHERE
			_key = ?
	`, e.Value, e.ExpireAt, e.Key)
	return err
}

func insertEntryWithTx(tx *sqlx.Tx, e *entry) error {
	res, err := tx.Exec(`
		INSERT INTO
			entries (_key, _value, _expire_at)
		VALUES
			(?, ?, ?)
	`, e.Key, e.Value, e.ExpireAt)
	if err != nil {
		return err
	}
	e.ID, _ = res.LastInsertId()
	return nil
}

func deleteEntryWithTx(tx *sqlx.Tx, key []byte) error {
	_, err := tx.Exec(`
		DELETE FROM
			entries
		WHERE
			_key = ?
	`, key)
	return err
}
