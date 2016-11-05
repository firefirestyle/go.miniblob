package handler

import (
	//	"net/url"

	"net/http"

	"github.com/firefirestyle/go.miniprop"
	"google.golang.org/appengine"
	"google.golang.org/appengine/blobstore"
)

func (obj *BlobHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	requestValues := r.URL.Query()
	key := requestValues.Get("key")
	dir := requestValues.Get("dir")
	file := requestValues.Get("file")

	obj.HandleGetBase(w, r, key, dir, file)
}

func (obj *BlobHandler) HandleGetBase(w http.ResponseWriter, r *http.Request, key, dir, file string) {

	//
	outputPropObj := miniprop.NewMiniProp()
	for _, f := range obj.onEvent.OnGetRequestList {
		errReqCheck := f(w, r, outputPropObj, obj)
		if errReqCheck != nil {
			HandleError(w, r, outputPropObj, ErrorCodeRequestCheck, errReqCheck.Error())
			return
		}
	}
	//
	if key != "" {
		w.Header().Set("Cache-Control", "public, max-age=2592000")
		for _, f := range obj.onEvent.OnGetSuccessList {
			f(w, r, outputPropObj, obj, nil)
		}
		blobstore.Send(w, appengine.BlobKey(key))
		return
	} else {
		ctx := appengine.NewContext(r)
		blobObj, err := obj.manager.GetBlobItemFromPointer(ctx, dir, file)
		if err != nil {
			for _, f := range obj.onEvent.OnGetFailedList {
				f(w, r, outputPropObj, obj, nil)
			}
			HandleError(w, r, outputPropObj, ErrorCodeGetBlobItem, err.Error())
			return
		} else {
			blobstore.Send(w, appengine.BlobKey(blobObj.GetBlobKey()))
			for _, f := range obj.onEvent.OnGetSuccessList {
				f(w, r, outputPropObj, obj, blobObj)
			}
			return
		}
	}
}
