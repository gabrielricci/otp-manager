package db

import (
	badger "github.com/dgraph-io/badger/v4"
	"log"
	"os"
)

var Connection *badger.DB

// OTPRepository Interface
type OTPRepository interface {
	SaveSecret(string, string) error
	GetSecret(string) (string, error)
}

// Badger Repository
type OTPBadgerRepository struct{}

func (repo *OTPBadgerRepository) SaveSecret(accountName string, secret string) (err error) {
	err = Connection.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(accountName), []byte(secret))
	})

	return
}

func (repo *OTPBadgerRepository) GetSecret(accountName string) (secret string, err error) {
	err = Connection.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(accountName))
		if err != nil {
			return err
		}

		copy, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		secret = string(copy)
		return nil
	})

	return
}

func Start() {
	db, err := badger.Open(badger.DefaultOptions(os.Getenv("DB_PATH")))
	if err != nil {
		log.Fatal(err)
	}

	Connection = db
}

func Stop() {
	Connection.Close()
}
