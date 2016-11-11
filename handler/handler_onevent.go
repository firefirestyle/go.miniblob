package handler

import (
	"net/http"

	"github.com/firefirestyle/go.miniprop"

	miniblob "github.com/firefirestyle/go.miniblob/blob"
	//	"golang.org/x/net/context"
	//	"google.golang.org/appengine/log"
)

func (obj *BlobHandler) AddOnBlobRequest(f func(w http.ResponseWriter, r *http.Request, input *miniprop.MiniProp, output *miniprop.MiniProp, h *BlobHandler) (map[string]string, error)) {
	obj.onEvent.OnBlobRequestList = append(obj.onEvent.OnBlobRequestList, f)
}

func (obj *BlobHandler) AddOnBlobBeforeSave(f func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem) error) {
	obj.onEvent.OnBlobBeforeSaveList = append(obj.onEvent.OnBlobBeforeSaveList, f)
}

func (obj *BlobHandler) AddOnBlobComplete(f func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem) error) {
	obj.onEvent.OnBlobCompleteList = append(obj.onEvent.OnBlobCompleteList, f)
}

func (obj *BlobHandler) AddOnBlobFailed(f func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem)) {
	obj.onEvent.OnBlobFailedList = append(obj.onEvent.OnBlobFailedList, f)
}

func (obj *BlobHandler) OnBlobRequestList(w http.ResponseWriter, r *http.Request, i *miniprop.MiniProp, o *miniprop.MiniProp, h *BlobHandler) (map[string]string, error) {
	ret := map[string]string{}
	for _, f := range obj.onEvent.OnBlobRequestList {
		vsTmp, err := f(w, r, i, o, h)
		for k, v := range vsTmp {
			ret[k] = v
		}
		if err != nil {
			return ret, err
		}
	}
	return ret, nil
}

func (obj *BlobHandler) OnBlobBeforeSave(w http.ResponseWriter, r *http.Request, o *miniprop.MiniProp, h *BlobHandler, i *miniblob.BlobItem) error {
	for _, f := range obj.onEvent.OnBlobBeforeSaveList {
		err := f(w, r, o, h, i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (obj *BlobHandler) OnBlobComplete(w http.ResponseWriter, r *http.Request, o *miniprop.MiniProp, h *BlobHandler, i *miniblob.BlobItem) error {
	for _, f := range obj.onEvent.OnBlobCompleteList {
		err := f(w, r, o, h, i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (obj *BlobHandler) OnBlobFailed(w http.ResponseWriter, r *http.Request, o *miniprop.MiniProp, h *BlobHandler, i *miniblob.BlobItem) {
	for _, f := range obj.onEvent.OnBlobFailedList {
		f(w, r, o, h, i)
	}
}

/**
 *
 * DeleteRequest
 *
 **/
func (obj *BlobHandler) AddOnDeleteRequest(f func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler) error) {
	obj.onEvent.OnDeleteRequestList = append(obj.onEvent.OnDeleteRequestList, f)
}

func (obj *BlobHandler) AddOnDeleteFailed(f func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem)) {
	obj.onEvent.OnDeleteFailedList = append(obj.onEvent.OnDeleteFailedList, f)
}

func (obj *BlobHandler) AddOnDeleteSuccess(f func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem)) {
	obj.onEvent.OnDeleteSuccessList = append(obj.onEvent.OnDeleteSuccessList, f)
}

func (obj *BlobHandler) OnDeleteRequest(w http.ResponseWriter, r *http.Request, o *miniprop.MiniProp, h *BlobHandler) error {
	for _, f := range obj.onEvent.OnDeleteRequestList {
		errReqCheck := f(w, r, o, h)
		if errReqCheck != nil {
			return errReqCheck
		}
	}
	return nil
}

func (obj *BlobHandler) OnDeleteFailed(w http.ResponseWriter, r *http.Request, o *miniprop.MiniProp, h *BlobHandler, i *miniblob.BlobItem) {
	for _, f := range obj.onEvent.OnDeleteFailedList {
		f(w, r, o, h, i)
	}
}

func (obj *BlobHandler) OnDeleteSuccess(w http.ResponseWriter, r *http.Request, o *miniprop.MiniProp, h *BlobHandler, i *miniblob.BlobItem) {
	for _, f := range obj.onEvent.OnDeleteSuccessList {
		f(w, r, o, h, i)
	}
}

/**
 *
 * GetRequest
 *
 **/
func (obj *BlobHandler) AddOnGetRequest(f func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler) error) {
	obj.onEvent.OnGetRequestList = append(obj.onEvent.OnGetRequestList, f)
}

func (obj *BlobHandler) AddOnGetFailed(f func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem)) {
	obj.onEvent.OnGetFailedList = append(obj.onEvent.OnGetFailedList, f)
}

func (obj *BlobHandler) AddOnGetSuccess(f func(http.ResponseWriter, *http.Request, *miniprop.MiniProp, *BlobHandler, *miniblob.BlobItem)) {
	obj.onEvent.OnGetSuccessList = append(obj.onEvent.OnGetSuccessList, f)
}

func (obj *BlobHandler) OnGetRequest(w http.ResponseWriter, r *http.Request, o *miniprop.MiniProp, h *BlobHandler) error {
	for _, f := range obj.onEvent.OnGetRequestList {
		errReqCheck := f(w, r, o, h)
		if errReqCheck != nil {
			return errReqCheck
		}
	}
	return nil
}

func (obj *BlobHandler) OnGetFailed(w http.ResponseWriter, r *http.Request, o *miniprop.MiniProp, h *BlobHandler, i *miniblob.BlobItem) {
	for _, f := range obj.onEvent.OnGetFailedList {
		f(w, r, o, h, i)
	}
}

func (obj *BlobHandler) OnGetSuccess(w http.ResponseWriter, r *http.Request, o *miniprop.MiniProp, h *BlobHandler, i *miniblob.BlobItem) {
	for _, f := range obj.onEvent.OnGetSuccessList {
		f(w, r, o, h, i)
	}
}
