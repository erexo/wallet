package main_test

import (
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/erexo/wallet/internal/domain"
	"github.com/erexo/wallet/internal/repository"
	"github.com/erexo/wallet/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSynchronousWallet(t *testing.T) {
	repo := repository.NewInMemory()
	walletService := service.New(repo)
	defer walletService.Close()

	id, err := walletService.Create()
	assert.NoError(t, err)

	funds, err := walletService.GetFunds(id)
	assert.NoError(t, err)
	assert.Equal(t, domain.Currency(0), funds)

	funds, err = walletService.ChangeFunds(id, -domain.FloatCurrency(.1))
	assert.Error(t, err)

	funds, err = walletService.ChangeFunds(id, domain.FloatCurrency(2.5))
	assert.NoError(t, err)
	assert.Equal(t, domain.Currency(250), funds)

	funds, err = walletService.ChangeFunds(id, -domain.FloatCurrency(3.))
	assert.Error(t, err)

	funds, err = walletService.ChangeFunds(id, domain.FloatCurrency(1.))
	assert.NoError(t, err)
	assert.Equal(t, domain.Currency(350), funds)
}

func TestAsynchronousWallet(t *testing.T) {
	repo := repository.NewInMemory()
	walletService := service.New(repo)
	defer walletService.Close()

	const (
		routines     = 1_000
		opPerRoutine = 1_000
		walletsCount = 10
	)

	wallets := make([]*wallet, 0, walletsCount)
	for i := 0; i < walletsCount; i++ {
		id, err := walletService.Create()
		assert.NoError(t, err)
		wallets = append(wallets, &wallet{id: id})
	}

	var wg sync.WaitGroup
	wg.Add(routines)

	for range routines {
		go func() {
			for range opPerRoutine {
				wallet := wallets[rand.Intn(len(wallets))]
				delta := rand.Int63n(56) - 20 // [-20,25]

				_, err := walletService.ChangeFunds(wallet.id, domain.Currency(delta))
				if errors.Is(err, service.ErrInsufficientFunds) {
					continue
				}
				assert.NoError(t, err)

				wallet.funds.Add(delta)
			}

			wg.Done()
		}()
	}

	wg.Wait()

	for _, wallet := range wallets {
		funds, err := walletService.GetFunds(wallet.id)
		assert.NoError(t, err)

		expectedFunds := domain.Currency(wallet.funds.Load())

		assert.Equal(t, expectedFunds, funds)
	}
}

func TestGracefulClose(t *testing.T) {
	expectedFunds := domain.Currency(100)

	repo := repository.NewDelayedInMemory(time.Second)
	walletService := service.New(repo)

	id, err := walletService.Create()
	assert.NoError(t, err)

	_, err = walletService.ChangeFunds(id, expectedFunds)
	assert.NoError(t, err)

	err = walletService.Close()
	assert.NoError(t, err)

	// service should not be usable after close
	_, err = walletService.ChangeFunds(id, expectedFunds)
	if assert.Error(t, err) {
		assert.Equal(t, service.ErrServiceIsClosed, err)
	}

	// service should not be closed twice
	err = walletService.Close()
	if assert.Error(t, err) {
		assert.Equal(t, service.ErrServiceIsClosed, err)
	}

	// opening new service on the same repository should contain the result from before the close
	walletService = service.New(repo)

	funds, err := walletService.GetFunds(id)
	if assert.NoError(t, err) {
		assert.Equal(t, expectedFunds, funds)
	}
}

func BenchmarkWalletFunds(b *testing.B) {
	repo := repository.NewInMemory()
	walletService := service.New(repo)
	defer walletService.Close()

	id, _ := walletService.Create()

	b.StartTimer()

	for n := 0; n < b.N; n++ {
		walletService.ChangeFunds(id, 100)
	}
}

func BenchmarkSingleWallet(b *testing.B) {
	repo := repository.NewInMemory()
	walletService := service.New(repo)
	defer walletService.Close()

	id1, _ := walletService.Create()

	b.StartTimer()

	for n := 0; n < b.N; n++ {
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			for range 100 {
				walletService.ChangeFunds(id1, 100)
			}
			wg.Done()
		}()
		go func() {
			for range 100 {
				walletService.ChangeFunds(id1, 100)
			}
			wg.Done()
		}()
		wg.Wait()
	}
}

func BenchmarkTwoWallets(b *testing.B) {
	repo := repository.NewInMemory()
	walletService := service.New(repo)
	defer walletService.Close()

	id1, _ := walletService.Create()
	id2, _ := walletService.Create()

	b.StartTimer()

	for n := 0; n < b.N; n++ {
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			for range 100 {
				walletService.ChangeFunds(id1, 100)
			}
			wg.Done()
		}()
		go func() {
			for range 100 {
				walletService.ChangeFunds(id2, 100)
			}
			wg.Done()
		}()
		wg.Wait()
	}
}

type wallet struct {
	id    uuid.UUID
	funds atomic.Int64
}
