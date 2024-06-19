package repository

import (
	"errors"

	"github.com/erexo/wallet/internal/domain"
	"github.com/google/uuid"
)

var (
	ErrWalletAlreadyExists = errors.New("wallet already exists")
	ErrWalletDoesNotExist  = errors.New("wallet does not exist")
)

type ChangeFunc func(current domain.Currency) (domain.Currency, error)

type Repository interface {
	Create(id uuid.UUID) error
	GetFunds(id uuid.UUID) (domain.Currency, error)
	ChangeFunds(id uuid.UUID, changeFunc ChangeFunc) (domain.Currency, error)
	Close() error
}
