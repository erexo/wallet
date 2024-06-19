package service

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/erexo/wallet/internal/domain"
	"github.com/erexo/wallet/internal/repository"
	"github.com/erexo/wallet/internal/utils"
	"github.com/google/uuid"
)

type Service interface {
	Create() (uuid.UUID, error)
	GetFunds(id uuid.UUID) (domain.Currency, error)
	ChangeFunds(id uuid.UUID, delta domain.Currency) (domain.Currency, error)
	Close() error
}

var (
	ErrServiceIsClosed   = errors.New("service is closed")
	ErrInsufficientFunds = errors.New("insufficient funds")
)

type service struct {
	repository repository.Repository
	locker     *utils.RWLocker[uuid.UUID]

	isClosed atomic.Bool
}

func New(repository repository.Repository) Service {
	return &service{
		repository: repository,
		locker:     utils.NewRWLocker[uuid.UUID](),
	}
}

func (s *service) Create() (uuid.UUID, error) {
	if s.isClosed.Load() {
		return uuid.Nil, ErrServiceIsClosed
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, errors.New("service: failed to generate uuid")
	}

	err = s.repository.Create(id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("service: failed to create '%v': %w", id, err)
	}

	return id, nil
}

func (s *service) GetFunds(id uuid.UUID) (domain.Currency, error) {
	if s.isClosed.Load() {
		return 0, ErrServiceIsClosed
	}

	unlock := s.locker.RLock(id)
	defer unlock()

	funds, err := s.repository.GetFunds(id)
	if err != nil {
		return 0, fmt.Errorf("service: failed to get funds '%v': %w", id, err)
	}

	return funds, nil
}

func (s *service) ChangeFunds(id uuid.UUID, delta domain.Currency) (domain.Currency, error) {
	if s.isClosed.Load() {
		return 0, ErrServiceIsClosed
	}

	unlock := s.locker.Lock(id)
	defer unlock()

	newFunds, err := s.repository.ChangeFunds(id, s.changeFunc(delta))
	if err != nil {
		return 0, fmt.Errorf("service: failed to change funds '%v' [%+d]: %w", id, delta, err)
	}

	return newFunds, nil
}

func (s *service) Close() error {
	if s.isClosed.Swap(true) {
		return ErrServiceIsClosed
	}

	s.locker.Wait()

	return nil
}

func (s *service) changeFunc(delta domain.Currency) repository.ChangeFunc {
	return func(current domain.Currency) (domain.Currency, error) {
		if -delta > current {
			return 0, ErrInsufficientFunds
		}

		return current + delta, nil
	}
}
