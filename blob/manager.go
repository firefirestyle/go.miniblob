package miniblob

import (
	"golang.org/x/net/context"
	//	"google.golang.org/appengine"
	"crypto/sha1"
	"errors"
	"net/http"
	"net/url"

	"io"

	"encoding/base64"
	"strings"

	"github.com/firefirestyle/go.minipointer"
	//	"github.com/firefirestyle/go.miniprop"
	"google.golang.org/appengine/blobstore"
	"google.golang.org/appengine/log"
)

type BlobManager struct {
	callbackUrl  string
	blobItemKind string
	rootGroup    string
	pointerMgr   *minipointer.PointerManager
}

type BlobManagerConfig struct {
	RootGroup   string
	Kind        string
	PointerKind string
	CallbackUrl string
}

func NewBlobManager(config BlobManagerConfig) *BlobManager {
	ret := new(BlobManager)
	ret.rootGroup = config.RootGroup
	ret.blobItemKind = config.Kind
	ret.callbackUrl = config.CallbackUrl
	ret.pointerMgr = minipointer.NewPointerManager(minipointer.PointerManagerConfig{
		RootGroup: config.RootGroup,
		Kind:      config.PointerKind,
	})
	return ret
}

func (obj *BlobManager) GetBlobItem(ctx context.Context, parent string, name string, sign string) (*BlobItem, error) {
	key := obj.NewBlobItemKey(ctx, parent, name, sign)
	Debug(ctx, "KEY ============="+key.StringID())

	return obj.NewBlobItemFromGaeObjectKey(ctx, key)
}

func (obj *BlobManager) GetBlobItemFromPointer(ctx context.Context, parent string, name string) (*BlobItem, error) {
	pointerObj, pointerErr := obj.pointerMgr.GetPointer(ctx, obj.GetBlobId(parent, name), minipointer.TypePointer)
	if pointerErr != nil {
		return nil, pointerErr
	}

	return obj.GetBlobItem(ctx, parent, name, pointerObj.GetSign())
}

func (obj *BlobManager) SaveBlobItemWithImmutable(ctx context.Context, newItem *BlobItem) error {
	errSave := newItem.saveDB(ctx)
	if errSave != nil {
		return errSave
	}

	currItem, currErr := obj.GetBlobItemFromPointer(ctx, newItem.GetParent(), newItem.GetName())
	pointerObj := obj.pointerMgr.GetPointerForRelayId(ctx, obj.GetBlobId(newItem.GetParent(), newItem.GetName()))

	pointerObj.SetSign(newItem.GetBlobKey())
	pointerObj.SetValue(newItem.gaeObjectKey.StringID())
	pointerErr := pointerObj.Save(ctx)
	if pointerErr != nil {
		err := newItem.deleteFromDB(ctx)
		if err != nil {
			Debug(ctx, "<gomidata>"+newItem.gaeObjectKey.StringID()+"</gomidata>")
		}
		return errSave
	}

	if currErr != nil {
		Debug(ctx, "===> SIGN A")
		return nil
	} else {
		Debug(ctx, "===> SIGN B")
		err := obj.DeleteBlobItem(ctx, currItem)
		if err != nil {
			Debug(ctx, "<gomidata>"+currItem.gaeObjectKey.StringID()+"</gomidata>")
		}
		return nil
	}
}

func (obj *BlobManager) DeleteBlobItem(ctx context.Context, item *BlobItem) error {
	return item.deleteFromDB(ctx)
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
	io.WriteString(hash, obj.rootGroup)
	io.WriteString(hash, dirName)
	io.WriteString(hash, obj.blobItemKind)
	io.WriteString(hash, fileName)
	io.WriteString(hash, privateSign)
	io.WriteString(hash, optKeyValue["kw"])

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
	io.WriteString(hash, obj.rootGroup)
	io.WriteString(hash, dirName)
	io.WriteString(hash, obj.blobItemKind)
	io.WriteString(hash, fileName)
	io.WriteString(hash, privateSign)
	io.WriteString(hash, r.FormValue("kw"))
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

func Debug(ctx context.Context, message string) {
	log.Infof(ctx, message)
}
