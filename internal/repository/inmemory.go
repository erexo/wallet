package repository

import (
	"sync"
	"time"

	"github.com/erexo/wallet/internal/domain"
	"github.com/google/uuid"
)

type inMemory struct {
	wallets  sync.Map
	setDelay time.Duration
}

func NewInMemory() Repository {
	return &inMemory{}
}

// ideally this should be mocked with ie. mockery
func NewDelayedInMemory(setDelay time.Duration) Repository {
	return &inMemory{
		setDelay: setDelay,
	}
}

func (r *inMemory) Create(id uuid.UUID) error {
	if _, loaded := r.wallets.LoadOrStore(id, domain.Currency(0)); loaded {
		return ErrWalletAlreadyExists
	}

	return nil
}

func (r *inMemory) GetFunds(id uuid.UUID) (domain.Currency, error) {
	if funds, ok := r.wallets.Load(id); ok {
		return funds.(domain.Currency), nil
	}

	return 0, ErrWalletDoesNotExist
}

func (r *inMemory) ChangeFunds(id uuid.UUID, changeFunc ChangeFunc) (domain.Currency, error) {
	if funds, ok := r.wallets.Load(id); ok {
		newFunds, err := changeFunc(funds.(domain.Currency))
		if err != nil {
			return 0, err
		}

		r.wallets.Store(id, newFunds)
		time.Sleep(r.setDelay)

		return newFunds, nil
	}

	return 0, ErrWalletDoesNotExist
}

func (r *inMemory) Close() error {
	return nil
}
