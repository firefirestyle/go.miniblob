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

	//	"github.com/firefirestyle/go.miniprop"
	"google.golang.org/appengine/blobstore"
)

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
