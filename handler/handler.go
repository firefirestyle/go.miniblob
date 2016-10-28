package handler

import (
	//	"net/url"

	"net/http"

	"github.com/firefirestyle/go.miniprop"

	miniblob "github.com/firefirestyle/go.miniblob/blob"
	"google.golang.org/appengine"
	//	"google.golang.org/appengine/blobstore"
)

type BlobHandlerOnEvent struct {
	OnRequest       func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler) (string, map[string]string, error)
	OnBeforeSave    func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem) error
	OnComplete      func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem) error
	OnFailed        func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem)
	OnDeleteRequest func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler) error
	OnDeleteFailed  func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem)
	OnDeleteSuccess func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem)
	OnGetRequest    func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler) error
	OnGetFailed     func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem)
	OnGetSuccess    func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem)
}

type BlobHandler struct {
	manager     *miniblob.BlobManager
	onEvent     BlobHandlerOnEvent
	callbackUrl string
	privateSign string
}

func (obj *BlobHandler) GetManager() *miniblob.BlobManager {
	return obj.manager
}

func NewBlobHandler(callbackUrl string, privateSign string, config miniblob.BlobManagerConfig, event BlobHandlerOnEvent) *BlobHandler {
	handlerObj := new(BlobHandler)
	handlerObj.privateSign = privateSign
	handlerObj.callbackUrl = callbackUrl
	handlerObj.manager = miniblob.NewBlobManager(config)
	handlerObj.onEvent = event
	if handlerObj.onEvent.OnRequest == nil {
		handlerObj.onEvent.OnRequest = func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler) (string, map[string]string, error) {
			return "dummy", map[string]string{}, nil
		}
	}
	if handlerObj.onEvent.OnComplete == nil {
		handlerObj.onEvent.OnComplete = func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem) error {
			return nil
		}
	}
	if handlerObj.onEvent.OnBeforeSave == nil {
		handlerObj.onEvent.OnBeforeSave = func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem) error {
			return nil
		}
	}
	if handlerObj.onEvent.OnFailed == nil {
		handlerObj.onEvent.OnFailed = func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem) {

		}
	}
	//
	if handlerObj.onEvent.OnDeleteRequest == nil {
		handlerObj.onEvent.OnDeleteRequest = func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler) error {
			return nil
		}
	}
	if handlerObj.onEvent.OnDeleteFailed == nil {
		handlerObj.onEvent.OnDeleteFailed = func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem) {
		}
	}
	if handlerObj.onEvent.OnDeleteSuccess == nil {
		handlerObj.onEvent.OnDeleteSuccess = func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem) {

		}
	}
	//
	if handlerObj.onEvent.OnGetRequest == nil {
		handlerObj.onEvent.OnGetRequest = func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler) error {
			return nil
		}
	}
	if handlerObj.onEvent.OnGetFailed == nil {
		handlerObj.onEvent.OnGetFailed = func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem) {
		}
	}
	if handlerObj.onEvent.OnGetSuccess == nil {
		handlerObj.onEvent.OnGetSuccess = func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem) {

		}
	}

	return handlerObj
}

func HandleError(w http.ResponseWriter, r *http.Request, outputProp *miniprop.MiniProp, errorCode int, errorMessage string) {
	//
	//
	if errorCode != 0 {
		outputProp.SetInt("errorCode", errorCode)
	}
	if errorMessage != "" {
		outputProp.SetString("errorMessage", errorMessage)
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Write(outputProp.ToJson())
}

const (
	ErrorCodeRequestCheck    = 2001
	ErrorCodeMakeRequestUrl  = 2002
	ErrorCodeCheckCallback   = 3001
	ErrorCodeBeforeSaveCheck = 3002
	ErrorCodeCompleteCheck   = 3003
	ErrorCodeSaveBlobItem    = 3004
	ErrorCodeGetBlobItem     = 3005
	ErrorCodeDeleteBlobItem  = 3006
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
	if obj.onEvent.OnRequest != nil {
		var err error = nil
		kv, vs, err = obj.onEvent.OnRequest(w, r, miniPropObj, obj)
		if err != nil {
			if obj.onEvent.OnFailed != nil {
				obj.onEvent.OnFailed(w, r, miniPropObj, obj, nil)
			}
			HandleError(w, r, miniPropObj, ErrorCodeRequestCheck, err.Error())
			return
		}
	}
	ctx := appengine.NewContext(r)
	uu, err := obj.manager.MakeRequestUrl(ctx, dirName, fileName, kv, obj.privateSign, vs)
	//
	if err != nil {
		if obj.onEvent.OnFailed != nil {
			obj.onEvent.OnFailed(w, r, miniPropObj, obj, nil)
		}
		HandleError(w, r, miniPropObj, ErrorCodeMakeRequestUrl, "failed to make uploadurl")
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
		if obj.onEvent.OnFailed != nil {
			obj.onEvent.OnFailed(w, r, miniPropObj, obj, nil)
		}
		HandleError(w, r, miniPropObj, ErrorCodeCheckCallback, e.Error())
		return
	}

	//
	ctx := appengine.NewContext(r)
	newItem := obj.manager.NewBlobItem(ctx, res.DirName, res.FileName, res.BlobKey)
	if obj.onEvent.OnBeforeSave != nil {
		err := obj.onEvent.OnBeforeSave(w, r, miniPropObj, obj, newItem)
		if err != nil {
			if obj.onEvent.OnFailed != nil {
				obj.onEvent.OnFailed(w, r, miniPropObj, obj, newItem)
			}
			HandleError(w, r, miniPropObj, ErrorCodeBeforeSaveCheck, "Failed to check")
			return
		}
	}
	err2 := obj.manager.SaveBlobItem(ctx, newItem)
	if err2 != nil {
		if obj.onEvent.OnFailed != nil {
			obj.onEvent.OnFailed(w, r, miniPropObj, obj, newItem)
		}
		HandleError(w, r, miniPropObj, ErrorCodeSaveBlobItem, "Failed to save blobitem")
		return
	}

	if obj.onEvent.OnComplete != nil {
		err := obj.onEvent.OnComplete(w, r, miniPropObj, obj, newItem)
		if err != nil {
			if obj.onEvent.OnFailed != nil {
				obj.onEvent.OnFailed(w, r, miniPropObj, obj, newItem)
			}
			HandleError(w, r, miniPropObj, ErrorCodeCompleteCheck, "Failed to save blobitem")
			return
		}
	}
	miniPropObj.SetString("blobkey", newItem.GetBlobKey())
	w.Write(miniPropObj.ToJson())
	w.WriteHeader(http.StatusOK)
}
