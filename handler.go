package miniblob

import (
	//	"net/url"

	"net/http"

	"github.com/firefirestyle/go.miniprop"

	"google.golang.org/appengine"
	"google.golang.org/appengine/blobstore"
)

type BlobHandlerOnEvent struct {
	OnRequest    func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler) (string, map[string]string, error)
	OnBeforeSave func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *BlobItem) error
	OnComplete   func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *BlobItem) error
	OnFailed     func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *BlobItem)
}

type BlobHandler struct {
	manager           *BlobManager
	onRequest         func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler) (string, map[string]string, error)
	onBeforeSave      func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *BlobItem) error
	onComplete        func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *BlobItem) error
	onFailedAtRequest func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *BlobItem)
	callbackUrl       string
	privateSign       string
}

func (obj *BlobHandler) GetManager() *BlobManager {
	return obj.manager
}

func NewBlobHandler(callbackUrl string, privateSign string, config BlobManagerConfig, event BlobHandlerOnEvent) *BlobHandler {
	handlerObj := new(BlobHandler)
	handlerObj.privateSign = privateSign
	handlerObj.callbackUrl = callbackUrl
	handlerObj.manager = NewBlobManager(config)
	handlerObj.onRequest = event.OnRequest
	handlerObj.onComplete = event.OnComplete
	handlerObj.onBeforeSave = event.OnBeforeSave
	handlerObj.onFailedAtRequest = event.OnFailed
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

const (
	ErrorCodeRequestCheck    = 2001
	ErrorCodeMakeRequestUrl  = 2002
	ErrorCodeCheckCallback   = 3001
	ErrorCodeBeforeSaveCheck = 3002
	ErrorCodeCompleteCheck   = 3003
	ErrorCodeSaveBlobItem    = 3004
)

func (obj *BlobHandler) HandleBlobRequestToken(w http.ResponseWriter, r *http.Request) {
	requestValues := r.URL.Query()
	dirName := requestValues.Get("dir")
	fileName := requestValues.Get("file")
	obj.HandleBlobRequestTokenFromParams(w, r, dirName, fileName)
}

func (obj *BlobHandler) HandleBlobRequestTokenFromParams(w http.ResponseWriter, r *http.Request, dirName string, fileName string) {
	miniPropObj := miniprop.NewMiniProp()
	//
	kv := "abcdef"
	vs := map[string]string{}
	if obj.onRequest != nil {
		var err error = nil
		kv, vs, err = obj.onRequest(w, r, miniPropObj, obj)
		if err != nil {
			miniPropObj.SetInt("errorCode", ErrorCodeRequestCheck)
			miniPropObj.SetString("errorMessage", err.Error())
			if obj.onFailedAtRequest != nil {
				obj.onFailedAtRequest(w, r, miniPropObj, obj, nil)
			}
			w.Write(miniPropObj.ToJson())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	ctx := appengine.NewContext(r)
	uu, err := obj.manager.MakeRequestUrl(ctx, dirName, fileName, kv, obj.privateSign, vs)
	//
	if err != nil {
		miniPropObj.SetInt("errorCode", ErrorCodeMakeRequestUrl)
		miniPropObj.SetString("errorMessage", "error://failed.to.make.uploadurl")
		if obj.onFailedAtRequest != nil {
			obj.onFailedAtRequest(w, r, miniPropObj, obj, nil)
		}
		w.Write(miniPropObj.ToJson())
		w.WriteHeader(http.StatusBadRequest)
	} else {
		miniPropObj.SetString("token", uu.String())
		w.Write(miniPropObj.ToJson())
		w.WriteHeader(http.StatusOK)
	}
}

func (obj *BlobHandler) HandleUploaded(w http.ResponseWriter, r *http.Request) {
	//
	miniPropObj := miniprop.NewMiniProp()
	res, e := obj.manager.CheckedCallback(r, obj.privateSign)
	if e != nil {
		miniPropObj.SetInt("errorCode", ErrorCodeCheckCallback)
		miniPropObj.SetString("errorMessage", e.Error())
		if obj.onFailedAtRequest != nil {
			obj.onFailedAtRequest(w, r, miniPropObj, obj, nil)
		}

		w.Write(miniPropObj.ToJson())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//
	ctx := appengine.NewContext(r)
	newItem := obj.manager.NewBlobItem(ctx, res.DirName, res.FileName, res.BlobKey)
	if obj.onBeforeSave != nil {
		err := obj.onBeforeSave(w, r, miniPropObj, obj, newItem)
		if err != nil {
			miniPropObj.SetInt("errorCode", ErrorCodeBeforeSaveCheck)
			miniPropObj.SetString("errorMessage", "Failed to check")
			if obj.onFailedAtRequest != nil {
				obj.onFailedAtRequest(w, r, miniPropObj, obj, newItem)
			}
			w.Write(miniPropObj.ToJson())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	err2 := obj.manager.SaveBlobItem(ctx, newItem)
	if err2 != nil {
		miniPropObj.SetInt("errorCode", ErrorCodeSaveBlobItem)
		miniPropObj.SetString("errorMessage", "Failed to save blobitem")
		if obj.onFailedAtRequest != nil {
			obj.onFailedAtRequest(w, r, miniPropObj, obj, newItem)
		}
		w.Write(miniPropObj.ToJson())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if obj.onComplete != nil {
		err := obj.onComplete(w, r, miniPropObj, obj, newItem)
		if err != nil {
			miniPropObj.SetInt("errorCode", ErrorCodeCompleteCheck)
			miniPropObj.SetString("errorMessage", "Failed to save blobitem")
			if obj.onFailedAtRequest != nil {
				obj.onFailedAtRequest(w, r, miniPropObj, obj, newItem)
			}
			w.Write(miniPropObj.ToJson())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	miniPropObj.SetString("blobkey", newItem.GetBlobKey())
	w.Write(miniPropObj.ToJson())
	w.WriteHeader(http.StatusOK)
}
