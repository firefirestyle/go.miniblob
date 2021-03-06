package blob

import (
	"golang.org/x/net/context"

	//	"time"

	//	"github.com/firefirestyle/go.miniprop"
	//	"google.golang.org/appengine"
	//	"google.golang.org/appengine/blobstore"
	"google.golang.org/appengine/datastore"
	//	"google.golang.org/appengine/memcache"
	"errors"

	"github.com/firefirestyle/go.minipointer"
)

func (obj *BlobManager) GetBlobItemFromGaeKey(ctx context.Context, gaeKey *datastore.Key) (*BlobItem, error) {
	memCacheObj, errMemCcache := obj.NewBlobItemFromMemcache(ctx, gaeKey.StringID())
	if errMemCcache == nil {
		return memCacheObj, nil
	}
	//
	//
	var item GaeObjectBlobItem
	err := datastore.Get(ctx, gaeKey, &item)
	if err != nil {
		return nil, err
	}
	ret := new(BlobItem)
	ret.gaeObject = &item
	ret.gaeKey = gaeKey

	if err == nil {
		ret.updateMemcache(ctx)
	}
	return ret, nil
}

func (obj *BlobManager) GetBlobItem(ctx context.Context, parent string, name string, sign string) (*BlobItem, error) {

	key := obj.NewBlobItemGaeKey(ctx, parent, name, sign)

	return obj.GetBlobItemFromGaeKey(ctx, key)
}

func (obj *BlobManager) GetBlobItemStringIdFromQuery(ctx context.Context, parent string, name string) (string, error) {
	founded := obj.FindBlobItemFromPath(ctx, parent, name, "")
	if len(founded.Keys) <= 0 {
		return "", errors.New("not found blobitem")
	}
	key := obj.NewBlobItemGaeKeyFromStringId(ctx, founded.Keys[0])
	return key.StringID(), nil
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
	pointerObj, pointerErr := obj.pointerMgr.GetPointer(ctx, obj.MakeBlobId(parent, name), minipointer.TypePointer)
	if pointerErr != nil {
		if obj.pointerMgr.IsMemcachedOnly() == false {
			return nil, nil, pointerErr
		} else {
			o, e := obj.GetBlobItemFromQuery(ctx, parent, name)
			var p *minipointer.Pointer = nil
			if e == nil {
				p, _ = obj.SavePointer(ctx, o)
			}
			return o, p, e
		}
	}
	retObj, retErr := obj.GetBlobItem(ctx, parent, name, pointerObj.GetSign())
	return retObj, pointerObj, retErr
}

func (obj *BlobManager) GetBlobItemStringIdFromPointer(ctx context.Context, parent string, name string) (string, string, error) {
	pointerObj, pointerErr := obj.pointerMgr.GetPointer(ctx, obj.MakeBlobId(parent, name), minipointer.TypePointer)
	if pointerErr != nil {
		if obj.pointerMgr.IsMemcachedOnly() == false {
			return "", "", pointerErr
		} else {
			o, e := obj.GetBlobItemStringIdFromQuery(ctx, parent, name)
			return o, obj.MakeBlobId(parent, name), e
		}
	}
	return pointerObj.GetValue(), obj.MakeBlobId(parent, name), nil
}
