package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/erexo/wallet/internal/domain"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type sqlite struct {
	db *sql.DB
}

func NewSQLite(path string) (Repository, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("sqlite: failed to open on path '%s': %w", path, err)
	}

	_, err = db.Exec("create table if not exists wallets (id blob not null primary key, funds int not null)")
	if err != nil {
		return nil, fmt.Errorf("sqlite: failed to create table '%s': %w", path, err)
	}

	return &sqlite{
		db: db,
	}, nil
}

func (s *sqlite) Create(id uuid.UUID) error {
	result, err := s.db.Exec("insert into wallets(id, funds) values(?, 0)", id)

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrWalletAlreadyExists
	}

	return nil
}

func (s *sqlite) GetFunds(id uuid.UUID) (domain.Currency, error) {
	var funds domain.Currency
	if err := s.db.QueryRow("select funds from wallets where id = ?", id).Scan(&funds); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrWalletDoesNotExist
		}

		return 0, err
	}

	return funds, nil
}

func (s *sqlite) ChangeFunds(id uuid.UUID, changeFunc ChangeFunc) (domain.Currency, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}

	// cashing current funds is also an option although more risky
	var funds domain.Currency
	// "for update" tx lock is not relevant in SQLite but could be useful in ie. MySQL
	if err := tx.QueryRow("select funds from wallets where id = ?", id).Scan(&funds); err != nil {
		tx.Rollback()

		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrWalletDoesNotExist
		}

		return 0, err
	}

	newFunds, err := changeFunc(funds)
	if err != nil {
		return 0, err
	}

	result, err := tx.Exec("update wallets set funds = ? where id = ?", newFunds, id)
	rows, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()

		return 0, err
	}

	if rows == 0 {
		return 0, ErrWalletDoesNotExist
	}

	return newFunds, tx.Commit()
}

func (s *sqlite) Close() error {
	return s.db.Close()
}
