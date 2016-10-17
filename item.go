package miniblob

import (
	"golang.org/x/net/context"

	"encoding/json"

	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/blobstore"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/memcache"
)

type GaeObjectBlobItem struct {
	ProjectId string
	Parent    string
	Name      string
	BlobKey   string
	Owner     string
	Info      string `datastore:",noindex"`
	Updated   time.Time
}

type BlobItem struct {
	gaeObject    *GaeObjectBlobItem
	gaeObjectKey *datastore.Key
}

const (
	TypeProjectId = "ProjectId"
	TypeParent    = "Parent"
	TypeName      = "Name"
	TypeBlobKey   = "BlobKey"
	TypeOwner     = "Owner"
	TypeInfo      = "Info"
	TypeUpdated   = "Updated"
)

func (obj *BlobItem) toJson() (string, error) {
	v := map[string]interface{}{
		TypeProjectId: obj.gaeObject.ProjectId,
		TypeParent:    obj.gaeObject.Parent,
		TypeName:      obj.gaeObject.Name,
		TypeBlobKey:   obj.gaeObjectKey.StringID(),
		TypeOwner:     obj.gaeObject.Owner,
		TypeInfo:      obj.gaeObject.Info,
		TypeUpdated:   obj.gaeObject.Updated.UnixNano(),
	}
	vv, e := json.Marshal(v)
	return string(vv), e
}

func (obj *BlobItem) updateMemcache(ctx context.Context) error {
	userObjMemSource, err_toJson := obj.toJson()
	if err_toJson == nil {
		userObjMem := &memcache.Item{
			Key:   obj.gaeObjectKey.StringID(),
			Value: []byte(userObjMemSource), //
		}
		memcache.Set(ctx, userObjMem)
	}
	return err_toJson
}

func (obj *BlobItem) SaveDB(ctx context.Context) error {
	_, e := datastore.Put(ctx, obj.gaeObjectKey, obj.gaeObject)
	obj.updateMemcache(ctx)
	return e
}

func (obj *BlobItem) DeleteFromDB(ctx context.Context) error {
	blobstore.Delete(ctx, appengine.BlobKey(obj.GetBlobKey()))
	return datastore.Delete(ctx, obj.gaeObjectKey)
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
