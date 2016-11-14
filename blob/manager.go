package blob

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
	config     BlobManagerConfig
	pointerMgr *minipointer.PointerManager
}

type BlobManagerConfig struct {
	RootGroup              string
	Kind                   string
	PointerKind            string
	CallbackUrl            string
	MemcachedOnlyInPointer bool
	HashLength             int
}

func NewBlobManager(config BlobManagerConfig) *BlobManager {
	ret := new(BlobManager)
	ret.config = config
	ret.pointerMgr = minipointer.NewPointerManager(minipointer.PointerManagerConfig{
		RootGroup:     config.RootGroup,
		Kind:          config.PointerKind,
		MemcachedOnly: config.MemcachedOnlyInPointer, // todo
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
