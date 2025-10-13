package jwtmgr

import "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/keystore"

func (m *RSManager) ListPublic() []keystore.PublicKey {
	if m == nil || m.ks == nil {
		return nil
	}
	return m.ks.ListPublic()
}

func (m *RSManager) Rotate(bits int) (string, error) {
	return m.ks.Rotate(bits)
}
