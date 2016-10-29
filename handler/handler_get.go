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

	//
	outputPropObj := miniprop.NewMiniProp()
	errReqCheck := obj.onEvent.OnGetRequest(w, r, outputPropObj, obj)
	if errReqCheck != nil {
		HandleError(w, r, outputPropObj, ErrorCodeRequestCheck, errReqCheck.Error())
		return
	}
	//
	if key != "" {
		w.Header().Set("Cache-Control", "public, max-age=2592000")
		obj.onEvent.OnGetSuccess(w, r, outputPropObj, obj, nil)
		blobstore.Send(w, appengine.BlobKey(key))
		return
	} else {
		ctx := appengine.NewContext(r)
		blobObj, err := obj.manager.GetBlobItem(ctx, dir, file)
		if err != nil {
			obj.onEvent.OnGetFailed(w, r, outputPropObj, obj, nil)
			HandleError(w, r, outputPropObj, ErrorCodeGetBlobItem, err.Error())
			return
		} else {
			blobstore.Send(w, appengine.BlobKey(blobObj.GetBlobKey()))
			obj.onEvent.OnGetSuccess(w, r, outputPropObj, obj, blobObj)
			return
		}
	}
}