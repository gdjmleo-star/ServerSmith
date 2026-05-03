package api

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"

	"golang.org/x/crypto/ssh"
)

// generateKeyPairBytes 生成 Ed25519 密钥对
// 返回 (OpenSSH 公钥行, OpenSSH PEM 私钥, error)
func generateKeyPairBytes() (pubKeyBytes []byte, privKeyBytes []byte, err error) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	// 公钥 → OpenSSH authorized_keys 格式
	sshPub, err := ssh.NewPublicKey(pubKey)
	if err != nil {
		return nil, nil, err
	}
	pubKeyBytes = ssh.MarshalAuthorizedKey(sshPub)

	// 私钥 → OpenSSH PEM 格式
	pemBlock, err := ssh.MarshalPrivateKey(privKey, "serversmith")
	if err != nil {
		return nil, nil, err
	}
	privKeyBytes = pem.EncodeToMemory(pemBlock)

	return pubKeyBytes, privKeyBytes, nil
}
