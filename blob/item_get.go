package miniblob

import (
	"golang.org/x/net/context"

	//	"time"

	//	"github.com/firefirestyle/go.miniprop"
	//	"google.golang.org/appengine"
	//	"google.golang.org/appengine/blobstore"
	"google.golang.org/appengine/datastore"
	//	"google.golang.org/appengine/memcache"
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
