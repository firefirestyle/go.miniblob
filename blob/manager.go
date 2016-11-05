package miniblob

import (
	"golang.org/x/net/context"
	//	"google.golang.org/appengine"

	"github.com/firefirestyle/go.minipointer"
	//	"github.com/firefirestyle/go.miniprop"
	"google.golang.org/appengine/log"
)

type BlobManager struct {
	callbackUrl  string
	blobItemKind string
	rootGroup    string
	pointerMgr   *minipointer.PointerManager
}

type BlobManagerConfig struct {
	RootGroup   string
	Kind        string
	PointerKind string
	CallbackUrl string
}

func NewBlobManager(config BlobManagerConfig) *BlobManager {
	ret := new(BlobManager)
	ret.rootGroup = config.RootGroup
	ret.blobItemKind = config.Kind
	ret.callbackUrl = config.CallbackUrl
	ret.pointerMgr = minipointer.NewPointerManager(minipointer.PointerManagerConfig{
		RootGroup: config.RootGroup,
		Kind:      config.PointerKind,
	})
	return ret
}

func (obj *BlobManager) GetBlobItem(ctx context.Context, parent string, name string, sign string) (*BlobItem, error) {
	key := obj.NewBlobItemKey(ctx, parent, name, sign)
	Debug(ctx, "KEY ============="+key.StringID())

	return obj.NewBlobItemFromGaeObjectKey(ctx, key)
}

func (obj *BlobManager) GetBlobItemFromPointer(ctx context.Context, parent string, name string) (*BlobItem, error) {
	pointerObj, pointerErr := obj.pointerMgr.GetPointer(ctx, obj.GetBlobId(parent, name), minipointer.TypePointer)
	if pointerErr != nil {
		return nil, pointerErr
	}

	return obj.GetBlobItem(ctx, parent, name, pointerObj.GetSign())
}

func (obj *BlobManager) SaveBlobItemWithImmutable(ctx context.Context, newItem *BlobItem) error {
	errSave := newItem.saveDB(ctx)
	if errSave != nil {
		return errSave
	}

	currItem, currErr := obj.GetBlobItemFromPointer(ctx, newItem.GetParent(), newItem.GetName())
	pointerObj := obj.pointerMgr.GetPointerForRelayId(ctx, obj.GetBlobId(newItem.GetParent(), newItem.GetName()))

	pointerObj.SetSign(newItem.GetBlobKey())
	pointerObj.SetValue(newItem.gaeObjectKey.StringID())
	pointerErr := pointerObj.Save(ctx)
	if pointerErr != nil {
		err := newItem.deleteFromDB(ctx)
		if err != nil {
			Debug(ctx, "<gomidata>"+newItem.gaeObjectKey.StringID()+"</gomidata>")
		}
		return errSave
	}

	if currErr != nil {
		Debug(ctx, "===> SIGN A")
		return nil
	} else {
		Debug(ctx, "===> SIGN B")
		err := obj.DeleteBlobItem(ctx, currItem)
		if err != nil {
			Debug(ctx, "<gomidata>"+currItem.gaeObjectKey.StringID()+"</gomidata>")
		}
		return nil
	}
}

func (obj *BlobManager) DeleteBlobItem(ctx context.Context, item *BlobItem) error {
	return item.deleteFromDB(ctx)
}

func Debug(ctx context.Context, message string) {
	log.Infof(ctx, message)
}
