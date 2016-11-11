package handler

import (
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
	//
	//
	reqErr := obj.OnDeleteRequest(w, r, outputPropObj, obj)
	if reqErr != nil {
		HandleError(w, r, outputPropObj, ErrorCodeAtDeleteRequestCheck, reqErr.Error())
		return
	}

	//
	//
	blolStringId, _, err := obj.manager.GetBlobItemStringIdFromPointer(ctx, dir, file)
	//	obj.manager.DeleteBlobItemWithPointerFromStringId(ctx, obj.manager.MakeStringId())
	//	blobObj, _, err := obj.manager.GetBlobItemFromPointer(ctx, dir, file)
	if err != nil {
		obj.OnDeleteFailed(w, r, outputPropObj, obj, nil)
		HandleError(w, r, outputPropObj, ErrorCodeAtDeleteRequestFindBlobItem, err.Error())
		return
	}
	errDelete := obj.manager.DeleteBlobItemWithPointerFromStringId(ctx, blolStringId)
	if errDelete != nil {
		obj.OnDeleteFailed(w, r, outputPropObj, obj, nil)
		HandleError(w, r, outputPropObj, ErrorCodeAtDeleteRequestDeleteBlobItem, err.Error())
		return
	} else {
		obj.OnDeleteSuccess(w, r, outputPropObj, obj, nil)
		w.Write(outputPropObj.ToJson())
		return
	}
}
