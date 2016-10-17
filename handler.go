package miniblob

import (
	//	"net/url"

	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/blobstore"
)

type BlobHandler struct {
	manager      *BlobManager
	onRequest    func(http.ResponseWriter, *http.Request, *BlobHandler) (string, map[string]string)
	onBeforeSave func(http.ResponseWriter, *http.Request, *BlobHandler, *BlobItem) error
	onComplete   func(http.ResponseWriter, *http.Request, *BlobHandler, *BlobItem) error
	onFailed     func(http.ResponseWriter, *http.Request, *BlobHandler, *BlobItem)
	callbackUrl  string
	privateSign  string
}

func (obj *BlobHandler) GetManager() *BlobManager {
	return obj.manager
}

func NewBlobHandler(callbackUrl string, privateSign string, //
	config BlobManagerConfig, //
	onRequest func(http.ResponseWriter, *http.Request, *BlobHandler) (string, map[string]string), //
	onBeforeSave func(http.ResponseWriter, *http.Request, *BlobHandler, *BlobItem) error,
	onComplete func(http.ResponseWriter, *http.Request, *BlobHandler, *BlobItem) error,
	onFailed func(http.ResponseWriter, *http.Request, *BlobHandler, *BlobItem)) *BlobHandler {
	handlerObj := new(BlobHandler)
	handlerObj.privateSign = privateSign
	handlerObj.callbackUrl = callbackUrl
	handlerObj.manager = NewBlobManager(config)
	handlerObj.onRequest = onRequest
	handlerObj.onComplete = onComplete
	handlerObj.onBeforeSave = onBeforeSave
	handlerObj.onFailed = onFailed
	return handlerObj
}

func (obj *BlobHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	requestValues := r.URL.Query()

	dir := requestValues.Get("dir")
	file := requestValues.Get("file")
	//
	ctx := appengine.NewContext(r)
	blobObj, err := obj.manager.GetBlobItem(ctx, dir, file)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		obj.manager.DeleteBlobItem(ctx, blobObj)
		return
	}
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
			w.Write([]byte(err.Error()))
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
	kv := "abcdef"
	vs := map[string]string{}
	if obj.onRequest != nil {
		kv, vs = obj.onRequest(w, r, obj)
	}
	ctx := appengine.NewContext(r)
	uu, err := obj.manager.MakeRequestUrl(ctx, dirName, fileName, kv, "", vs)
	//
	if err != nil {
		w.Write([]byte("error://failed.to.make.uploadurl"))
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.Write([]byte(uu.String()))
	}
}

func (obj *BlobHandler) HandleUploaded(w http.ResponseWriter, r *http.Request) {
	//
	res, e := obj.manager.CheckedCallback(r, "")
	if e != nil {
		w.Write([]byte(e.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//
	ctx := appengine.NewContext(r)
	newItem := obj.manager.NewBlobItem(ctx, res.DirName, res.FileName, res.BlobKey)
	if obj.onBeforeSave != nil {
		err := obj.onBeforeSave(w, r, obj, newItem)
		if err != nil {
			w.Write([]byte("Failed to save blobitem"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	err2 := obj.manager.SaveBlobItem(ctx, newItem)
	if err2 != nil {
		if obj.onFailed != nil {
			obj.onFailed(w, r, obj, newItem)
		}
		w.Write([]byte("Failed to save blobitem"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if obj.onComplete != nil {
		err := obj.onComplete(w, r, obj, newItem)
		if err != nil {
			w.Write([]byte("Failed to save blobitem"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	w.Write([]byte(newItem.GetBlobKey()))
	w.WriteHeader(http.StatusOK)
}
