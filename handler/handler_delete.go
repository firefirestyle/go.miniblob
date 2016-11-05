package handler

import (
	//	"net/url"

	"net/http"

	"github.com/firefirestyle/go.miniprop"
	"google.golang.org/appengine"
)

func (obj *BlobHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	outputPropObj := miniprop.NewMiniProp()
	requestValues := r.URL.Query()
	dir := requestValues.Get("dir")
	file := requestValues.Get("file")
	ctx := appengine.NewContext(r)
	{
		for _, ff := range obj.onEvent.OnDeleteRequest {
			err := ff(w, r, outputPropObj, obj)
			if err != nil {
				HandleError(w, r, outputPropObj, ErrorCodeRequestCheck, err.Error())
				return
			}
		}
	}
	blobObj, err := obj.manager.GetBlobItemFromPointer(ctx, dir, file)
	if err != nil {
		for _, ff := range obj.onEvent.OnDeleteFailed {
			ff(w, r, outputPropObj, obj, nil)
		}
		HandleError(w, r, outputPropObj, ErrorCodeGetBlobItem, err.Error())
		return
	} else {
		errDelete := obj.manager.DeleteBlobItem(ctx, blobObj)
		if errDelete != nil {
			for _, ff := range obj.onEvent.OnDeleteFailed {
				ff(w, r, outputPropObj, obj, blobObj)
			}
			HandleError(w, r, outputPropObj, ErrorCodeDeleteBlobItem, err.Error())
		} else {
			for _, f := range obj.onEvent.OnDeleteSuccess {
				f(w, r, outputPropObj, obj, blobObj)
			}
		}
		return
	}
}
