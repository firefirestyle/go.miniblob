package miniblob

import (
	"golang.org/x/net/context"
	//	"google.golang.org/appengine"

	//	"errors"

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

func (obj *BlobManager) GetPointer(ctx context.Context, parent, name string) (*minipointer.Pointer, error) {
	return obj.pointerMgr.GetPointer(ctx, obj.MakeBlobId(parent, name), minipointer.TypePointer)
}

func (obj *BlobManager) GetPointerGaeKey(ctx context.Context, parent, name string) *datastore.Key {
	return obj.pointerMgr.NewPointerGaeKey(ctx, obj.MakeBlobId(parent, name), minipointer.TypePointer)
}

func Debug(ctx context.Context, message string) {
	log.Infof(ctx, message)
}
