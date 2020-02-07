package lego

import (
	"github.com/ihaiker/aginx/storage"
	"time"
)

type Manager struct {
	AccountStorage     *AccountStorage
	CertificateStorage *CertificateStorage
	ticker             *time.Ticker
	OnDomainExpire     func(domain string)
}

func NewManager(engine storage.Engine) (manager *Manager, err error) {
	manager = new(Manager)
	if manager.AccountStorage, err = LoadAccounts(engine); err != nil {
		return
	}
	if manager.CertificateStorage, err = LoadCertificates(engine); err != nil {
		return
	}
	manager.ticker = time.NewTicker(time.Second * 5)
	return
}

func (manager *Manager) Start() error {
	go func() {
		for {
			select {
			case <-manager.ticker.C:
				for domain, certificate := range manager.CertificateStorage.data {
					if certificate.ExpireTime.Before(time.Now()) {
						if manager.OnDomainExpire != nil {
							manager.OnDomainExpire(domain)
						}
					}
				}
			}
		}
	}()
	return nil
}

func (manager *Manager) Stop() error {
	if manager.ticker != nil {
		manager.ticker.Stop()
	}
	return nil
}
