package miniblob

import (
	"golang.org/x/net/context"
	//	"google.golang.org/appengine"
	//"google.golang.org/appengine/blobstore"
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

func (obj *BlobManager) SaveBlobItem(ctx context.Context, newItem *BlobItem) error {
	oldItem, err2 := obj.GetBlobItem(ctx, newItem.GetParent(), newItem.GetName())
	if err2 == nil {
		oldItem.deleteFromDB(ctx)
	}
	return newItem.saveDB(ctx)
}

func (obj *BlobManager) DeleteBlobItem(ctx context.Context, item *BlobItem) error {
	return item.deleteFromDB(ctx)
}

func (obj *BlobManager) MakeStringId(parent string, name string) string {
	return "" + obj.projectId + "://" + parent + "/" + name
}
