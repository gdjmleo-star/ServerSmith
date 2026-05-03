package api

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

const (
	keyDir      = "/var/lib/serversmith/keys"
	privKeyPath = "/var/lib/serversmith/keys/serversmith_ed25519"
	pubKeyPath  = "/var/lib/serversmith/keys/serversmith_ed25519.pub"
)

// ensureKeyPair 确保 Ed25519 密钒对存在，返回公钒字符串
func ensureKeyPair() (string, error) {
	if pubBytes, err := os.ReadFile(pubKeyPath); err == nil {
		return string(pubBytes), nil
	}
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return "", err
	}
	pubKey, privKey, err := generateKeyPairBytes()
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(privKeyPath, privKey, 0600); err != nil {
		return "", err
	}
	if err := os.WriteFile(pubKeyPath, pubKey, 0644); err != nil {
		return "", err
	}
	log.Printf("[ssh-key] Generated new Ed25519 key pair at %s", keyDir)
	return strings.TrimSpace(string(pubKey)), nil
}

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
