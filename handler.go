package miniblob

import (
	"strings"

	"net/url"

	"encoding/base64"

	"golang.org/x/net/context"

	"net/http"

	"errors"

	"google.golang.org/appengine"
	"google.golang.org/appengine/blobstore"

	"crypto/sha1"
	"io"
)

type BlobHandler struct {
	manager     *BlobManager
	onRequest   func(url.Values) (string, map[string]string)
	onComplete  func(url.Values, *BlobItem) map[string]string
	callbackUrl string
}

func NewBlobHandler(callbackUrl string, //
	config BlobManagerConfig, //
	onRequest func(url.Values) (string, map[string]string), //
	onComplete func(url.Values, *BlobItem) map[string]string) *BlobHandler {
	handlerObj := new(BlobHandler)
	handlerObj.callbackUrl = callbackUrl
	handlerObj.manager = NewBlobManager(config)
	handlerObj.onRequest = onRequest
	handlerObj.onComplete = onComplete
	return handlerObj
}

func (obj *BlobHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	requestValues := r.URL.Query()
	key := requestValues.Get("key")
	dir := requestValues.Get("dir")
	file := requestValues.Get("file")

	//
	if key != "" {
		w.Header().Set("Cache-Control", "public, max-age=2592000")
		blobstore.Send(w, appengine.BlobKey(key))
		return
	} else {
		ctx := appengine.NewContext(r)
		blobObj, err := obj.manager.GetBlobItem(ctx, dir, file)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else {
			blobstore.Send(w, appengine.BlobKey(blobObj.GetBlobKey()))
			return
		}
	}
}

func (obj *BlobHandler) BlobRequestToken(w http.ResponseWriter, r *http.Request) {
	requestValues := r.URL.Query()
	dirName := requestValues.Get("dir")
	fileName := requestValues.Get("file")
	//
	//
	callbackUrlObj, _ := url.Parse(obj.callbackUrl)
	callbackValue := callbackUrlObj.Query()
	callbackValue.Add("dir", dirName)
	callbackValue.Add("file", fileName)
	//
	hash := sha1.New()
	io.WriteString(hash, obj.manager.projectId)
	io.WriteString(hash, dirName)
	io.WriteString(hash, obj.manager.blobItemKind)
	io.WriteString(hash, fileName)

	//
	if obj.onRequest != nil {
		kv, vs := obj.onRequest(r.URL.Query())
		callbackValue.Add("kv", kv)
		io.WriteString(hash, kv)
		for k, v := range vs {
			callbackValue.Add(k, v)
		}
	}
	callbackValue.Add("hash", base64.StdEncoding.EncodeToString(hash.Sum(nil)))
	callbackUrlObj.RawQuery = callbackValue.Encode()
	//
	//
	ctx := appengine.NewContext(r)
	uu, err := blobstore.UploadURL(ctx, callbackUrlObj.String(), nil)
	//
	//
	if err != nil {
		w.Write([]byte("error://failed.to.make.uploadurl"))
	} else {
		w.Write([]byte(uu.String()))
	}
}

func (obj *BlobHandler) HandleUploaded(w http.ResponseWriter, r *http.Request) {
	//
	blobs, _, err := blobstore.ParseUpload(r)
	if err != nil {
		// error
		w.Write([]byte("Failed to make blobls"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hashValue := r.FormValue("hash")
	dirName := r.FormValue("dir")
	fileName := r.FormValue("file")
	kv := r.FormValue("kv")
	hash := sha1.New()
	io.WriteString(hash, obj.manager.projectId)
	io.WriteString(hash, dirName)
	io.WriteString(hash, obj.manager.blobItemKind)
	io.WriteString(hash, fileName)
	if kv != "" {
		io.WriteString(hash, kv)
	}
	calcHash := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	if 0 != strings.Compare(calcHash, hashValue) {
		w.Write([]byte("Failed to make hash"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// --
	// files
	// --
	file := blobs["file"]
	if len(file) == 0 {
		w.Write([]byte("Failed to make files"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	//
	// opt
	blobKey := string(file[0].BlobKey)
	if fileName == "" {
		fileName = blobKey
	}

	//
	//
	//
	ctx := appengine.NewContext(r)
	newItem := obj.manager.NewBlobItem(ctx, dirName, fileName, blobKey)

	err2 := obj.manager.SaveBlobItem(ctx, newItem)
	if err2 != nil {
		w.Write([]byte("Failed to save blobitem"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if obj.onComplete != nil {
		obj.onComplete(r.URL.Query(), newItem)
	}

}

//
//
//
//
//
//
func (obj *BlobManager) MakeRequestUrl(ctx context.Context, dirName string, fileName string, opt string) (string, error) {
	if opt == "" {
		opt = "none"
	}

	var ary = []string{obj.callbackUrl + //
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
	//
	blobs, _, err := blobstore.ParseUpload(r)
	if err != nil {
		return nil, "", err
	}

	// --
	// dirName
	// --
	dirNameSrc, err1 := base64.StdEncoding.DecodeString(r.FormValue("dir"))
	if err1 != nil {
		return nil, "", err1
	}
	dirName := string(dirNameSrc)

	// --
	// filename
	// --
	fileName := r.FormValue("file")

	// --
	// opt
	// --
	optProp := string(r.FormValue("opt"))

	// --
	// file
	// --
	file := blobs["file"]
	if len(file) == 0 {
		return nil, "", errors.New("")
	}
	blobKey := string(file[0].BlobKey)
	if fileName == "" {
		fileName = blobKey
	}

	//
	//
	//
	newItem := obj.NewBlobItem(ctx, dirName, fileName, blobKey)
	err2 := obj.SaveBlobItem(ctx, newItem)
	if err2 != nil {
		return nil, "", errors.New("")
	}
	return newItem, optProp, err
}
