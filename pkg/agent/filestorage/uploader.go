package filestorage

import "context"

type Uploader interface {
	UploadUrl(ctx context.Context, fileUrl string) (string, error)
	UploadJson(ctx context.Context, json interface{}) (string, error)
}
