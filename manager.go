package miniblob

import (
	"golang.org/x/net/context"

	"encoding/json"

	"time"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
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

func (obj *BlobManager) NewBlobItemFromMemcache(ctx context.Context, keyId string) (*BlobItem, error) {
	jsonSource, errGetJsonSource := memcache.Get(ctx, keyId)
	if errGetJsonSource != nil {
		return nil, errGetJsonSource
	}
	v := make(map[string]interface{})
	e := json.Unmarshal(jsonSource.Value, &v)
	if e != nil {
		return nil, e
	}

	ret := new(BlobItem)
	ret.gaeObjectKey = datastore.NewKey(ctx, obj.blobItemKind, keyId, 0, nil)
	ret.gaeObject = new(GaeObjectBlobItem)
	ret.gaeObject.ProjectId = v[TypeProjectId].(string)
	ret.gaeObject.Parent = v[TypeParent].(string)
	ret.gaeObject.Name = v[TypeName].(string)
	ret.gaeObject.BlobKey = v[TypeBlobKey].(string)
	ret.gaeObject.Info = v[TypeInfo].(string)
	ret.gaeObject.Updated = time.Unix(0, int64(v[TypeUpdated].(float64)))

	return ret, nil
}

func NewBlobManager(config BlobManagerConfig) *BlobManager {
	ret := new(BlobManager)
	ret.projectId = config.ProjectId
	ret.blobItemKind = config.Kind
	ret.BasePath = config.UrlRoot
	return ret
}

func (obj *BlobManager) GetBlobItem(ctx context.Context, parent string, name string, isNew bool) (*BlobItem, error) {
	key := obj.NewBlobItemKey(ctx, parent, name)
	return obj.NewBlobItemFromGaeObjectKey(ctx, key)
}

func (obj *BlobManager) MakeStringId(parent string, name string) string {
	return "" + obj.projectId + "://" + parent + "/" + name
}

func (obj *BlobManager) NewBlobItemKey(ctx context.Context, parent string, name string) *datastore.Key {
	return datastore.NewKey(ctx, obj.blobItemKind, obj.MakeStringId(parent, name), 0, nil)
}

func (obj *BlobManager) NewBlobItem(ctx context.Context, parent string, name string, blobKey string) *BlobItem {
	ret := new(BlobItem)
	ret.gaeObject = new(GaeObjectBlobItem)
	ret.gaeObject.ProjectId = obj.projectId
	ret.gaeObject.Parent = parent
	ret.gaeObject.Name = name
	ret.gaeObject.BlobKey = blobKey
	ret.gaeObject.Updated = time.Now()
	ret.gaeObjectKey = datastore.NewKey(ctx, obj.blobItemKind, ""+parent+"/"+name, 0, nil)
	return ret
}

func (obj *BlobManager) NewBlobItemFromGaeObjectKey(ctx context.Context, gaeKey *datastore.Key) (*BlobItem, error) {
	memCacheObj, errMemCcache := obj.NewBlobItemFromMemcache(ctx, gaeKey.StringID())
	if errMemCcache == nil {
		log.Infof(ctx, ">>>> from memcache "+obj.projectId+":"+gaeKey.StringID())
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
	ret.gaeObjectKey = gaeKey

	if err == nil {
		ret.updateMemcache(ctx)
	}
	return ret, nil
}
