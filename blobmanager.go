package miniblob

import (
	"strings"

	"net/url"

	"encoding/base64"

	"golang.org/x/net/context"

	"bytes"
	"net/http"

	"errors"

	"encoding/json"
	"mime/multipart"

	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/blobstore"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
	"google.golang.org/appengine/urlfetch"
)

type BlobManager struct {
	BasePath     string
	blobItemKind string
	projectId    string
}

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
	var item GaeObjectBlobItem
	err := datastore.Get(ctx, key, &item)

	var ret *BlobItem = nil
	if err != nil {
		if isNew == true {
			item.Name = name
			item.Parent = parent
			ret = obj.NewBlobItemFromGaeObject(ctx, key, &item)
		} else {
			return nil, err
		}
	} else {
		ret = obj.NewBlobItemFromGaeObject(ctx, key, &item)
	}
	ret.updateMemcache(ctx)
	return ret, nil
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

func (obj *BlobManager) NewBlobItemFromGaeObject(ctx context.Context, gaeKey *datastore.Key, gaeObj *GaeObjectBlobItem) *BlobItem {
	ret := new(BlobItem)
	ret.gaeObject = gaeObj
	ret.gaeObjectKey = gaeKey
	return ret
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
	founds := q.Run(ctx)

	var retUser []*BlobItem

	var cursorNext string = ""
	var cursorOne string = ""

	for i := 0; ; i++ {
		var d GaeObjectBlobItem
		key, err := founds.Next(&d)
		if err != nil || err == datastore.Done {
			break
		} else {
			retUser = append(retUser, obj.NewBlobItemFromGaeObject(ctx, key, &d))
		}
		if i == 0 {
			cursorOne = obj.makeCursorSrc(founds)
		}
	}
	cursorNext = obj.makeCursorSrc(founds)
	return retUser, cursorOne, cursorNext
}

//
//
//
func (obj *BlobManager) MakeRequestUrl(ctx context.Context, dirName string, fileName string, opt string) (string, error) {
	if opt == "" {
		opt = "none"
	}

	var ary = []string{obj.BasePath + //
		"?dir=", url.QueryEscape(base64.StdEncoding.EncodeToString([]byte(dirName))), //
		"&file=", url.QueryEscape(fileName), //
		"&opt=", opt}
	uu, err2 := blobstore.UploadURL(ctx, strings.Join(ary, ""), nil) //&option)
	return uu.String(), err2
}

//
//
//
func (obj *BlobManager) HandleUploaded(ctx context.Context, r *http.Request) (*BlobItem, string, error) {
	blobs, _, err := blobstore.ParseUpload(r)
	if err != nil {
		// error
		return nil, "", err
	}
	dirNameSrc, err1 := base64.StdEncoding.DecodeString(r.FormValue("dir"))
	if err1 != nil {
		// error
		return nil, "", err1
	}
	dirName := string(dirNameSrc)
	fileName := r.FormValue("file")

	reqId := string(r.FormValue("opt"))

	file := blobs["file"]
	if len(file) == 0 {
		// error
		return nil, "", errors.New("")
	}
	blobKey := string(file[0].BlobKey)
	if fileName == "" {
		fileName = blobKey
	}
	blobItem, err2 := obj.GetBlobItem(ctx, dirName, fileName, false)
	if err2 == nil {
		blobstore.Delete(ctx, appengine.BlobKey(blobItem.GetBlobKey()))
		blobItem.gaeObject.BlobKey = blobKey
	} else {
		blobItem = obj.NewBlobItem(ctx, dirName, fileName, blobKey)
	}

	err = blobItem.SaveDB(ctx)
	if err != nil {
		blobstore.Delete(ctx, appengine.BlobKey(blobKey))
	}
	return blobItem, reqId, err
}

func (obj *BlobManager) SaveData(c context.Context, url string, sampleData []byte) error {

	// Now you can prepare a form that you will submit to that URL.
	var b bytes.Buffer
	fw := multipart.NewWriter(&b)
	// Do not change the form field, it must be "file"!
	// You are free to change the filename though, it will be stored in the BlobInfo.
	file, err := fw.CreateFormFile("file", "example.csv")
	if err != nil {
		return err
	}
	if _, err = file.Write(sampleData); err != nil {
		return err
	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	fw.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return err
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", fw.FormDataContentType())

	// Now submit the request.
	client := urlfetch.Client(c)
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	// Check the response status, it should be whatever you return in the `/upload` handler.
	if res.StatusCode != http.StatusCreated {
		return err
	}
	// Everything went fine.
	return nil
}

//

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
