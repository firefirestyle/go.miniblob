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

func NewBlobManager(projectId string, uploadUrlBase string, blobItemKind string) *BlobManager {
	ret := new(BlobManager)
	ret.projectId = projectId
	ret.blobItemKind = blobItemKind
	ret.BasePath = uploadUrlBase
	return ret
}

func (obj *BlobManager) GetBlobItem(ctx context.Context, parent string, name string, isNew bool) (*BlobItem, error) {
	memCacheObj, errMemCcache := obj.NewBlobItemFromMemcache(ctx, obj.MakeStringId(parent, name))
	if errMemCcache == nil {
		log.Infof(ctx, ">>>> from memcache "+obj.projectId+"://"+parent+"/"+name)
		return memCacheObj, nil
	}
	key := obj.NewBlobItemKey(ctx, parent, name)
	ret, err := obj.NewBlobItemFromGaeObjectKey(ctx, key)

	if err == nil {
		ret.updateMemcache(ctx)
		return ret, nil
	} else {
		return nil, err
	}
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
	var item GaeObjectBlobItem
	err := datastore.Get(ctx, gaeKey, &item)
	if err != nil {
		return nil, err
	}
	ret := new(BlobItem)
	ret.gaeObject = &item
	ret.gaeObjectKey = gaeKey
	return ret, nil
}

/*
func (obj *BlobManager) NewBlobItemFromGaeObject(ctx context.Context, gaeKey *datastore.Key, gaeObj *GaeObjectBlobItem) *BlobItem {
	ret := new(BlobItem)
	ret.gaeObject = gaeObj
	ret.gaeObjectKey = gaeKey
	return ret
}
*/
/*
- kind: BlobItem
  properties:
  - name: Parent
  - name: Updated
    direction: asc

- kind: BlobItem
  properties:
  - name: Parent
  - name: Updated
    direction: desc
https://cloud.google.com/appengine/docs/go/config/indexconfig#updating_indexes
*/
func (obj *BlobManager) FindBlobItemFromParent(ctx context.Context, parent string, cursorSrc string) ([]*BlobItem, string, string) {
	//
	q := datastore.NewQuery(obj.blobItemKind)
	q = q.Filter("ProjectId =", obj.projectId)
	q = q.Filter("Parent =", parent)
	q = q.Order("-Updated")
	//
	return obj.FindBlobItemFromQuery(ctx, q, cursorSrc)
}

//
//
func (obj *BlobManager) FindBlobItemFromQuery(ctx context.Context, q *datastore.Query, cursorSrc string) ([]*BlobItem, string, string) {
	cursor := obj.newCursorFromSrc(cursorSrc)
	if cursor != nil {
		q = q.Start(*cursor)
	}
	q = q.KeysOnly()
	founds := q.Run(ctx)

	var retUser []*BlobItem

	var cursorNext string = ""
	var cursorOne string = ""

	for i := 0; ; i++ {
		key, err := founds.Next(nil)
		if err != nil || err == datastore.Done {
			break
		} else {
			v, e := obj.NewBlobItemFromGaeObjectKey(ctx, key)
			if e == nil {
				retUser = append(retUser, v)
			}
		}
		if i == 0 {
			cursorOne = obj.makeCursorSrc(founds)
		}
	}
	cursorNext = obj.makeCursorSrc(founds)
	return retUser, cursorOne, cursorNext
}

func (obj *BlobManager) newCursorFromSrc(cursorSrc string) *datastore.Cursor {
	c1, e := datastore.DecodeCursor(cursorSrc)
	if e != nil {
		return nil
	} else {
		return &c1
	}
}

func (obj *BlobManager) makeCursorSrc(founds *datastore.Iterator) string {
	c, e := founds.Cursor()
	if e == nil {
		return c.String()
	} else {
		return ""
	}
}
