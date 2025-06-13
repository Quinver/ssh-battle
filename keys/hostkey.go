package keys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"

	gossh "golang.org/x/crypto/ssh"
)

func LoadHostKey(path string) (gossh.Signer, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, err
		}
		keyBytes := x509.MarshalPKCS1PrivateKey(key)
		pemKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyBytes})
		if err := os.WriteFile(path, pemKey, 0600); err != nil {
			return nil, err
		}
	}

	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return gossh.ParsePrivateKey(keyBytes)
}
