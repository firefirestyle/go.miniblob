package miniblob

import (
	"golang.org/x/net/context"

	"time"

	"github.com/firefirestyle/go.miniprop"
	"google.golang.org/appengine"
	"google.golang.org/appengine/blobstore"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/memcache"
)

type GaeObjectBlobItem struct {
	RootGroup string
	Parent    string
	Name      string
	BlobKey   string
	Owner     string
	Info      string `datastore:",noindex"`
	Updated   time.Time
	Sign      string `datastore:",noindex"`
}

type BlobItem struct {
	gaeObject    *GaeObjectBlobItem
	gaeObjectKey *datastore.Key
}

const (
	TypeRootGroup = "RootGroup"
	TypeParent    = "Parent"
	TypeName      = "Name"
	TypeBlobKey   = "BlobKey"
	TypeOwner     = "Owner"
	TypeInfo      = "Info"
	TypeUpdated   = "Updated"
	TypeSign      = "Sign"
)

func (obj *BlobManager) NewBlobItem(ctx context.Context, parent string, name string, blobKey string) *BlobItem {
	ret := new(BlobItem)
	ret.gaeObject = new(GaeObjectBlobItem)
	ret.gaeObject.RootGroup = obj.rootGroup
	ret.gaeObject.Parent = parent
	ret.gaeObject.Name = name
	ret.gaeObject.BlobKey = blobKey
	ret.gaeObject.Updated = time.Now()
	ret.gaeObject.Sign = blobKey
	ret.gaeObjectKey = datastore.NewKey(ctx, obj.blobItemKind, obj.MakeStringId(parent, name, blobKey), 0, nil)
	return ret
}

func (obj *BlobManager) NewBlobItemFromMemcache(ctx context.Context, keyId string) (*BlobItem, error) {
	jsonSource, errGetJsonSource := memcache.Get(ctx, keyId)
	if errGetJsonSource != nil {
		return nil, errGetJsonSource
	}

	ret := new(BlobItem)
	ret.gaeObjectKey = datastore.NewKey(ctx, obj.blobItemKind, keyId, 0, nil)
	ret.gaeObject = new(GaeObjectBlobItem)
	err := ret.SetParamFromJson(jsonSource.Value)
	return ret, err
}

func (obj *BlobManager) NewBlobItemKey(ctx context.Context, parent string, name string, sign string) *datastore.Key {
	return datastore.NewKey(ctx, obj.blobItemKind, obj.MakeStringId(parent, name, sign), 0, nil)
}

func (obj *BlobManager) NewBlobItemFromGaeObjectKey(ctx context.Context, gaeKey *datastore.Key) (*BlobItem, error) {
	memCacheObj, errMemCcache := obj.NewBlobItemFromMemcache(ctx, gaeKey.StringID())
	if errMemCcache == nil {
		Debug(ctx, ">>>> from memcache "+obj.rootGroup+":"+gaeKey.StringID())
		return memCacheObj, nil
	}
	//
	//
	var item GaeObjectBlobItem
	err := datastore.Get(ctx, gaeKey, &item)
	if err != nil {
		Debug(ctx, ">>>> failed to get "+obj.rootGroup+":"+gaeKey.StringID())
		return nil, err
	}
	Debug(ctx, ">>>> from datastore to get "+obj.rootGroup+":"+gaeKey.StringID())
	ret := new(BlobItem)
	ret.gaeObject = &item
	ret.gaeObjectKey = gaeKey

	if err == nil {
		ret.updateMemcache(ctx)
	}
	return ret, nil
}

func (obj *BlobItem) updateMemcache(ctx context.Context) error {
	userObjMemSource, err_toJson := obj.ToJson()
	if err_toJson == nil {
		userObjMem := &memcache.Item{
			Key:   obj.gaeObjectKey.StringID(),
			Value: []byte(userObjMemSource), //
		}
		memcache.Set(ctx, userObjMem)
	}
	return err_toJson
}

func (obj *BlobItem) saveDB(ctx context.Context) error {
	_, e := datastore.Put(ctx, obj.gaeObjectKey, obj.gaeObject)
	obj.updateMemcache(ctx)
	return e
}

func (obj *BlobItem) deleteFromDB(ctx context.Context) error {
	Debug(ctx, "delete From DB OLD ITEM =A============GK"+obj.gaeObjectKey.StringID()+";BK:"+obj.GetBlobKey())
	if nil != blobstore.Delete(ctx, appengine.BlobKey(obj.GetBlobKey())) {
		Debug(ctx, "SaveBlobItem Faied Blob: "+obj.gaeObjectKey.StringID()+":"+obj.GetBlobKey())
	}
	return datastore.Delete(ctx, obj.gaeObjectKey)
}

func (obj *BlobManager) MakeStringId(parent string, name string, sign string) string {
	propObj := miniprop.NewMiniProp()
	propObj.SetString("p", obj.rootGroup)
	propObj.SetString("d", parent)
	propObj.SetString("f", name)
	propObj.SetString("s", sign)
	return string(propObj.ToJson())
}

func (obj *BlobManager) GetBlobId(parent string, name string) string {
	propObj := miniprop.NewMiniProp()
	propObj.SetString("p", obj.rootGroup)
	propObj.SetString("d", parent)
	propObj.SetString("f", name)
	return string(propObj.ToJson())
}

func (obj *BlobItem) GetParent() string {
	return obj.gaeObject.Parent
}

func (obj *BlobItem) GetName() string {
	return obj.gaeObject.Name
}

func (obj *BlobItem) GetBlobKey() string {
	return obj.gaeObject.BlobKey
}

func (obj *BlobItem) GetInfo() string {
	return obj.gaeObject.Info
}

func (obj *BlobItem) SetInfo(v string) {
	obj.gaeObject.Info = v
}

func (obj *BlobItem) GetSign() string {
	return obj.gaeObject.Sign
}

/*func (obj *BlobItem) SetBlobKey(v string) {
	obj.gaeObject.BlobKey = v
}*/
