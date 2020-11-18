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
	expireFunc         func(domain string)
	engine             plugins.StorageEngine
}

func NewManager(engine plugins.StorageEngine) (manager *Manager, err error) {
	manager = new(Manager)
	manager.ticker = time.NewTicker(time.Minute)
	manager.engine = engine
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
func (manager *Manager) Start() (err error) {
	if manager.AccountStorage, err = LoadAccounts(manager.engine); err != nil {
		return
	}
	if manager.CertificateStorage, err = LoadCertificates(manager.engine); err != nil {
		return
	}
	go func() {
		logrus.Debug("start check certificate expire")
		for {
			select {
			case <-manager.ticker.C:
				logrus.Debug("check certificate expire")
				for domain, certificate := range manager.CertificateStorage.data {
					if certificate.IsExpire(time.Minute) {
						logrus.Infof("%s expire，re apply\n", domain)
						manager.applyForACertificate(domain)
					}
				}
			}
		}
	}()
	return
}

func (manager *Manager) Stop() error {
	if manager.ticker != nil {
		manager.ticker.Stop()
	}
	return nil
}
