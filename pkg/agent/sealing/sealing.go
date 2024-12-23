package sealing

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"

	"github.com/Dstack-TEE/dstack/sdk/go/tappd"

	"github.com/NethermindEth/yayois-garden/pkg/agent/debug"
)

func getSealingKey(ctx context.Context, dstackTappdEndpoint string) ([]byte, error) {
	dstackTappdClient := tappd.NewTappdClient(tappd.WithEndpoint(dstackTappdEndpoint))

	sealingKeyResp, err := dstackTappdClient.DeriveKeyWithSubject(ctx, "/agent/sealing", "teeception")
	if err != nil {
		return nil, fmt.Errorf("failed to derive sealing key: %v", err)
	}

	sealingKey, err := sealingKeyResp.ToBytes(32)
	if err != nil {
		return nil, fmt.Errorf("failed to convert sealing key to bytes: %v", err)
	}

	return sealingKey, nil
}

func WriteSealedFile(ctx context.Context, dstackTappdEndpoint, filePath string, data []byte) error {
	if debug.IsDebugPlainSetup() {
		return writeFilePlain(filePath, data)
	}

	return writeFileSealed(ctx, dstackTappdEndpoint, filePath, data)
}

func ReadSealedFile(ctx context.Context, dstackTappdEndpoint, filePath string) ([]byte, error) {
	if debug.IsDebugPlainSetup() {
		return readFilePlain(filePath)
	}

	return readFileSealed(ctx, dstackTappdEndpoint, filePath)
}

func writeFilePlain(filePath string, data []byte) error {
	return os.WriteFile(filePath, data, 0600)
}

func readFilePlain(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

func writeFileSealed(ctx context.Context, dstackTappdEndpoint, filePath string, data []byte) error {
	key, err := getSealingKey(ctx, dstackTappdEndpoint)
	if err != nil {
		return fmt.Errorf("failed to get sealing key: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %v", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to create nonce: %v", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	if err := os.WriteFile(filePath, ciphertext, 0600); err != nil {
		return fmt.Errorf("failed to write secure file: %v", err)
	}

	return nil
}

func readFileSealed(ctx context.Context, dstackTappdEndpoint, filePath string) ([]byte, error) {
	key, err := getSealingKey(ctx, dstackTappdEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get sealing key: %v", err)
	}

	ciphertext, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read secure file: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %v", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %v", err)
	}

	return plaintext, nil
}
