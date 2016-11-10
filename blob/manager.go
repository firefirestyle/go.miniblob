package miniblob

import (
	"golang.org/x/net/context"
	//	"google.golang.org/appengine"

	"errors"

	"github.com/firefirestyle/go.minipointer"
	//	"github.com/firefirestyle/go.miniprop"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type BlobManager struct {
	callbackUrl  string
	blobItemKind string
	rootGroup    string
	pointerMgr   *minipointer.PointerManager
}

type BlobManagerConfig struct {
	RootGroup     string
	Kind          string
	PointerKind   string
	CallbackUrl   string
	MemcachedOnly bool
}

func NewBlobManager(config BlobManagerConfig) *BlobManager {
	ret := new(BlobManager)
	ret.rootGroup = config.RootGroup
	ret.blobItemKind = config.Kind
	ret.callbackUrl = config.CallbackUrl
	ret.pointerMgr = minipointer.NewPointerManager(minipointer.PointerManagerConfig{
		RootGroup:     config.RootGroup,
		Kind:          config.PointerKind,
		MemcachedOnly: config.MemcachedOnly, // todo
	})
	return ret
}

func (obj *BlobManager) GetPointerMgr() *minipointer.PointerManager {

	return obj.pointerMgr
}

func (obj *BlobManager) GetBlobItem(ctx context.Context, parent string, name string, sign string) (*BlobItem, error) {

	key := obj.NewBlobItemGaeKey(ctx, parent, name, sign)

	return obj.GetBlobItemFromGaeKey(ctx, key)
}

func (obj *BlobManager) GetBlobItemFromQuery(ctx context.Context, parent string, name string) (*BlobItem, error) {
	founded := obj.FindBlobItemFromPath(ctx, parent, name, "")
	if len(founded.Keys) <= 0 {
		return nil, errors.New("not found blobitem")
	}
	key := obj.NewBlobItemGaeKeyFromStringId(ctx, founded.Keys[0])
	return obj.GetBlobItemFromGaeKey(ctx, key)
}

func (obj *BlobManager) GetBlobItemFromStringId(ctx context.Context, stringId string) (*BlobItem, error) {
	key := obj.NewBlobItemGaeKeyFromStringId(ctx, stringId)
	return obj.GetBlobItemFromGaeKey(ctx, key)
}

//
// if memcachedonly == true , posssible to become pointer == null
func (obj *BlobManager) GetBlobItemFromPointer(ctx context.Context, parent string, name string) (*BlobItem, *minipointer.Pointer, error) {
	pointerObj, pointerErr := obj.pointerMgr.GetPointer(ctx, obj.GetBlobId(parent, name), minipointer.TypePointer)
	if pointerErr != nil {
		if obj.pointerMgr.IsMemcachedOnly() == false {
			return nil, nil, pointerErr
		} else {
			o, e := obj.GetBlobItemFromQuery(ctx, parent, name)
			return o, nil, e
		}
	}
	retObj, retErr := obj.GetBlobItem(ctx, parent, name, pointerObj.GetSign())
	return retObj, pointerObj, retErr
}

func (obj *BlobManager) GetPointer(ctx context.Context, parent, name string) (*minipointer.Pointer, error) {
	return obj.pointerMgr.GetPointer(ctx, obj.GetBlobId(parent, name), minipointer.TypePointer)
}

func (obj *BlobManager) GetPointerGaeKey(ctx context.Context, parent, name string) *datastore.Key {
	return obj.pointerMgr.NewPointerGaeKey(ctx, obj.GetBlobId(parent, name), minipointer.TypePointer)
}

func Debug(ctx context.Context, message string) {
	log.Infof(ctx, message)
}
