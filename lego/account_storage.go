package lego

import (
	"encoding/json"
	"github.com/go-acme/lego/v3/certcrypto"
	"github.com/go-acme/lego/v3/lego"
	"github.com/go-acme/lego/v3/registration"
	"github.com/ihaiker/aginx/plugins"
)

const accountDir = "lego/accounts"

type AccountStorage struct {
	store  map[string]*Account
	engine plugins.StorageEngine
}

func (acs *AccountStorage) Get(email string) (account *Account, has bool) {
	account, has = acs.store[email]
	return
}

func (acs *AccountStorage) restore(email string) error {
	account, _ := acs.Get(email)
	file := accountDir + "/" + email + ".json"
	bs, err := json.MarshalIndent(account, "", "\t")
	if err != nil {
		return err
	}
	if err := acs.engine.Put(file, bs); err != nil {
		return err
	}
	return nil
}

func (acs *AccountStorage) registration(account *Account) error {
	config := lego.NewConfig(account)
	config.Certificate.KeyType = account.KeyType

	client, err := lego.NewClient(config)
	if err != nil {
		return err
	}

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return err
	}
	account.Registration = reg
	return nil
}

func (acs *AccountStorage) New(email string, keyType certcrypto.KeyType) (*Account, error) {
	if account, has := acs.Get(email); has {
		return account, nil
	}

	privateKey, err := certcrypto.GeneratePrivateKey(keyType)
	if err != nil {
		return nil, err
	}

	account := &Account{Email: email, KeyType: keyType}
	if err := account.SetKey(privateKey); err != nil {
		return nil, err
	}

	if err = acs.registration(account); err != nil {
		return nil, err
	}

	acs.store[email] = account
	err = acs.restore(email)
	return account, err
}

func LoadAccounts(engine plugins.StorageEngine) (accountStorage *AccountStorage, err error) {
	accountStorage = &AccountStorage{
		store: map[string]*Account{}, engine: engine,
	}

	files, err := engine.Search(accountDir + "/*.json")
	if err != nil {
		return
	}

	for _, file := range files {
		path := file.Name
		keyBytes := file.Content

		account := new(Account)
		err = json.Unmarshal(keyBytes, account)
		if err == nil {
			accountStorage.store[account.Email] = account
		}

		logrus.WithError(err).Info("load account file ", path)
	}
	return
}
