package handler

import (
	//	"net/url"

	"net/http"

	"github.com/firefirestyle/go.miniprop"

	miniblob "github.com/firefirestyle/go.miniblob/blob"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	//	"google.golang.org/appengine"
	//	"google.golang.org/appengine/blobstore"
	//	"errors"
)

type BlobHandlerOnEvent struct {
	OnBlobRequest    []func(w http.ResponseWriter, r *http.Request, input *miniprop.MiniProp, output *miniprop.MiniProp, h *BlobHandler) (string, map[string]string, error)
	OnBlobBeforeSave []func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem) error
	OnBlobComplete   []func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem) error
	OnBlobFailed     []func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem)
	OnDeleteRequest  []func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler) error
	OnDeleteFailed   []func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem)
	OnDeleteSuccess  []func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem)
	OnGetRequest     []func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler) error
	OnGetFailed      []func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem)
	OnGetSuccess     []func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem)
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

func NewBlobHandler(callbackUrl string, privateSign string, config miniblob.BlobManagerConfig) *BlobHandler {
	handlerObj := new(BlobHandler)
	handlerObj.privateSign = privateSign
	handlerObj.callbackUrl = callbackUrl
	handlerObj.manager = miniblob.NewBlobManager(config)
	handlerObj.onEvent = BlobHandlerOnEvent{}
	return handlerObj
}

func (obj *BlobHandler) GetBlobHandleEvent() *BlobHandlerOnEvent {
	return &obj.onEvent
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

func Debug(ctx context.Context, message string) {
	log.Infof(ctx, message)
}
