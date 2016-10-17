package miniblob

import (
	"golang.org/x/net/context"
	//	"google.golang.org/appengine"
	"crypto/sha1"
	"net/http"
	"net/url"

	"errors"

	"io"

	"encoding/base64"
	"strings"

	"google.golang.org/appengine/blobstore"
)

type BlobManager struct {
	callbackUrl  string
	blobItemKind string
	projectId    string
}

type BlobManagerConfig struct {
	ProjectId   string
	Kind        string
	CallbackUrl string
}

func NewBlobManager(config BlobManagerConfig) *BlobManager {
	ret := new(BlobManager)
	ret.projectId = config.ProjectId
	ret.blobItemKind = config.Kind
	ret.callbackUrl = config.CallbackUrl
	return ret
}

func (obj *BlobManager) GetBlobItem(ctx context.Context, parent string, name string) (*BlobItem, error) {
	key := obj.NewBlobItemKey(ctx, parent, name)
	return obj.NewBlobItemFromGaeObjectKey(ctx, key)
}

func (obj *BlobManager) SaveBlobItem(ctx context.Context, newItem *BlobItem) error {
	oldItem, err2 := obj.GetBlobItem(ctx, newItem.GetParent(), newItem.GetName())
	if err2 == nil {
		oldItem.deleteFromDB(ctx)
	}
	return newItem.saveDB(ctx)
}

func (obj *BlobManager) DeleteBlobItem(ctx context.Context, item *BlobItem) error {
	return item.deleteFromDB(ctx)
}

func (obj *BlobManager) MakeStringId(parent string, name string) string {
	return "" + obj.projectId + "://" + parent + "/" + name
}

//
// for make original hundler
//

func (obj *BlobManager) MakeRequestUrl(ctx context.Context, dirName string, fileName string, publicSign string, privateSign string, optKeyValue map[string]string) (*url.URL, error) {
	//
	//
	callbackUrlObj, _ := url.Parse(obj.callbackUrl)
	callbackValue := callbackUrlObj.Query()
	callbackValue.Add("dir", dirName)
	callbackValue.Add("file", fileName)
	//
	hash := sha1.New()
	io.WriteString(hash, obj.projectId)
	io.WriteString(hash, dirName)
	io.WriteString(hash, obj.blobItemKind)
	io.WriteString(hash, fileName)
	io.WriteString(hash, privateSign)

	//
	callbackValue.Add("kv", publicSign)
	io.WriteString(hash, publicSign)
	if optKeyValue != nil {
		for k, v := range optKeyValue {
			callbackValue.Add(k, v)
		}
	}
	callbackValue.Add("hash", base64.StdEncoding.EncodeToString(hash.Sum(nil)))
	callbackUrlObj.RawQuery = callbackValue.Encode()
	return blobstore.UploadURL(ctx, callbackUrlObj.String(), nil)
}

type CheckCallbackResult struct {
	DirName  string
	FileName string
	BlobKey  string
}

func (obj *BlobManager) CheckedCallback(r *http.Request, privateSign string) (*CheckCallbackResult, error) {
	//
	blobs, _, err := blobstore.ParseUpload(r)
	if err != nil {
		return nil, errors.New("faied to parseupload")
	}

	hashValue := r.FormValue("hash")
	dirName := r.FormValue("dir")
	fileName := r.FormValue("file")
	kv := r.FormValue("kv")
	hash := sha1.New()
	io.WriteString(hash, obj.projectId)
	io.WriteString(hash, dirName)
	io.WriteString(hash, obj.blobItemKind)
	io.WriteString(hash, fileName)
	io.WriteString(hash, privateSign)
	if kv != "" {
		io.WriteString(hash, kv)
	}
	calcHash := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	if 0 != strings.Compare(calcHash, hashValue) {
		return nil, errors.New("faied to check hash")
	}

	// --
	// files
	// --
	file := blobs["file"]
	if len(file) == 0 {
		return nil, errors.New("faied to find file")
	}
	//
	// opt
	blobKey := string(file[0].BlobKey)
	if fileName == "" {
		fileName = blobKey
	}

	return &CheckCallbackResult{
		DirName:  dirName,
		FileName: fileName,
		BlobKey:  blobKey,
	}, nil
}
