package registry

import "github.com/ihaiker/aginx/plugins"

type MultiRegister []plugins.Register

func (m *MultiRegister) Add(register plugins.Register) {
	*m = append(*m, register)
}

func (m MultiRegister) Size() int {
	return len(m)
}

func (m MultiRegister) Start() error {
	for _, register := range m {
		if err := register.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (m MultiRegister) Stop() error {
	for _, register := range m {
		_ = register.Stop()
	}
	return nil
}
