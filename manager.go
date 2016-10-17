package miniblob

import (
	"golang.org/x/net/context"
)

type BlobManager struct {
	BasePath     string
	blobItemKind string
	projectId    string
}

type BlobManagerConfig struct {
	ProjectId string
	Kind      string
	UrlRoot   string
}

func NewBlobManager(config BlobManagerConfig) *BlobManager {
	ret := new(BlobManager)
	ret.projectId = config.ProjectId
	ret.blobItemKind = config.Kind
	ret.BasePath = config.UrlRoot
	return ret
}

func (obj *BlobManager) GetBlobItem(ctx context.Context, parent string, name string) (*BlobItem, error) {
	key := obj.NewBlobItemKey(ctx, parent, name)
	return obj.NewBlobItemFromGaeObjectKey(ctx, key)
}

func (obj *BlobManager) MakeStringId(parent string, name string) string {
	return "" + obj.projectId + "://" + parent + "/" + name
}
