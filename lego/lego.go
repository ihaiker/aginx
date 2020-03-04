package lego

import (
	"github.com/ihaiker/aginx/plugins"
	"github.com/ihaiker/aginx/util"
	"time"
)

type Manager struct {
	AccountStorage     *AccountStorage
	CertificateStorage *CertificateStorage
	ticker             *time.Ticker

	expireFunc func(domain string)
}

func NewManager(engine plugins.StorageEngine) (manager *Manager, err error) {
	manager = new(Manager)
	if manager.AccountStorage, err = LoadAccounts(engine); err != nil {
		return
	}
	if manager.CertificateStorage, err = LoadCertificates(engine); err != nil {
		return
	}
	manager.ticker = time.NewTicker(time.Hour)
	return
}

func (manager *Manager) Expire(expireFunc func(domain string)) {
	manager.expireFunc = expireFunc
}

func (manager *Manager) applyForACertificate(domain string) {
	defer util.Catch(func(err error) {
		logrus.Warnf("Request for %s certificate exception: %s ", domain, err)
	})
	if manager.expireFunc != nil {
		manager.expireFunc(domain)
	}
}
func (manager *Manager) Start() error {
	go func() {
		for {
			select {
			case <-manager.ticker.C:
				for domain, certificate := range manager.CertificateStorage.data {
					if certificate.ExpireTime.Before(time.Now().Add(time.Hour)) {
						manager.applyForACertificate(domain)
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
