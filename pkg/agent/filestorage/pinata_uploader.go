package filestorage

import (
	"context"
	"fmt"

	"github.com/zde37/pinata-go-sdk/pinata"
)

type PinataUploader struct {
	jwtKey string

	client *pinata.Client
}

func NewPinataUploader(jwtKey string) *PinataUploader {
	return &PinataUploader{
		jwtKey: jwtKey,
		client: pinata.New(pinata.NewAuthWithJWT(jwtKey)),
	}
}

func (u *PinataUploader) UploadUrl(ctx context.Context, fileUrl string) (string, error) {
	pinResponse, err := u.client.PinURL(fileUrl, nil)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to pinata: %v", err)
	}

	return pinResponse.IpfsHash, nil
}

func (u *PinataUploader) UploadJson(ctx context.Context, json interface{}) (string, error) {
	pinResponse, err := u.client.PinJSON(json, nil)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to pinata: %v", err)
	}

	return pinResponse.IpfsHash, nil
}
