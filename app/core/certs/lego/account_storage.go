package lego

import (
	"encoding/json"
	"github.com/go-acme/lego/v3/certcrypto"
	"github.com/go-acme/lego/v3/lego"
	"github.com/go-acme/lego/v3/registration"
	"github.com/ihaiker/aginx/v2/api"
)

type accountStorage struct {
	accountDir string
	store      map[string]*Account
	aginx      api.Aginx
}

func (acs *accountStorage) Get(email string) (account *Account, has bool) {
	account, has = acs.store[email]
	return
}

func (acs *accountStorage) restore(email string) error {
	account, _ := acs.Get(email)
	file := acs.accountDir + "/" + email + ".json"
	bs, err := json.MarshalIndent(account, "", "\t")
	if err != nil {
		return err
	}
	if err := acs.aginx.Files().NewWithContent(file, bs); err != nil {
		return err
	}
	return nil
}

func (acs *accountStorage) registration(account *Account) error {
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

func (acs *accountStorage) New(email string, keyType certcrypto.KeyType) (*Account, error) {
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

func loadAccounts(baseDir string, aginx api.Aginx) (as *accountStorage, err error) {
	as = &accountStorage{
		store: map[string]*Account{}, aginx: aginx,
		accountDir: baseDir,
	}

	files, err := aginx.Files().Search(baseDir + "/*.json")
	if err != nil {
		return
	}

	for _, file := range files {
		path := file.Name
		keyBytes := file.Content

		account := new(Account)
		err = json.Unmarshal(keyBytes, account)
		if err == nil {
			as.store[account.Email] = account
			logger.Debug("加载账户文件 ", path)
		} else {
			logger.WithError(err).Warn("加载账户文件 ", path)
		}
	}
	return
}
