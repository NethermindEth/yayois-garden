package nft

import (
	"context"
	"fmt"

	"github.com/NethermindEth/yayois-garden/pkg/agent/filestorage"
)

type NftUploader struct {
	uploader filestorage.Uploader
}

func NewNftUploader(uploader filestorage.Uploader) *NftUploader {
	return &NftUploader{
		uploader: uploader,
	}
}

func (u *NftUploader) Upload(ctx context.Context, name, description, imageUri string) (string, error) {
	imageIpfsHash, err := u.uploader.UploadUrl(ctx, imageUri)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to ipfs: %v", err)
	}

	metadataIpfsHash, err := u.uploader.UploadJson(ctx, map[string]string{
		"name":        name,
		"description": description,
		"image":       imageIpfsHash,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to ipfs: %v", err)
	}

	return metadataIpfsHash, nil
}
