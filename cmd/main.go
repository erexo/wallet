package main

import (
	"fmt"

	"github.com/erexo/wallet/internal/domain"
	"github.com/erexo/wallet/internal/repository"
	"github.com/erexo/wallet/internal/service"
	"github.com/google/uuid"
)

/* todos:
   config
   logs
   mocks
*/

func main() {
	dbfile := "./wallets.db" // ideally in some sort of config

	if err := run(dbfile); err != nil {
		panic(err)
	}
}

func run(dbfile string) error {
	repo, err := repository.NewSQLite(dbfile)
	if err != nil {
		return err
	}
	defer repo.Close()

	service := service.New(repo)
	defer service.Close()

	// other services should interact here with Service's interface via topic, queue, gRPC, REST or any other form of communication

	return mockDial(service)
}

func mockDial(service service.Service) error {
	idstr := "" // already created wallet id or empty

	id, err := uuid.Parse(idstr)
	if err != nil {
		id, err = service.Create()
		if err != nil {
			return err
		}
	}

	fmt.Println("Using id:", id)

	funds, err := service.GetFunds(id)
	if err != nil {
		return err
	}

	fmt.Println("Wallet holds", funds, "funds")

	addFunds := domain.FloatCurrency(3.50)
	newFunds, err := service.ChangeFunds(id, addFunds)
	if err != nil {
		return err
	}

	fmt.Printf("Added %v funds to the wallet, current funds: %v\n", addFunds, newFunds)

	return nil
}
