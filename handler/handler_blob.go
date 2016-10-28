package handler

import (
	//	"net/url"

	"net/http"

	"github.com/firefirestyle/go.miniprop"

	//	miniblob "github.com/firefirestyle/go.miniblob/blob"
	"google.golang.org/appengine"
	//	"google.golang.org/appengine/blobstore"
)

func (obj *BlobHandler) HandleBlobRequestToken(w http.ResponseWriter, r *http.Request) {
	requestValues := r.URL.Query()
	dirName := requestValues.Get("dir")
	fileName := requestValues.Get("file")
	obj.HandleBlobRequestTokenFromParams(w, r, dirName, fileName)
}

func (obj *BlobHandler) HandleBlobRequestTokenFromParams(w http.ResponseWriter, r *http.Request, dirName string, fileName string) {
	ctx := appengine.NewContext(r)
	outputPropObj := miniprop.NewMiniProp()
	//
	kv := "abcdef"
	vs := map[string]string{}
	{
		var err error = nil
		kv, vs, err = obj.onEvent.OnBlobRequest(w, r, outputPropObj, obj)
		if err != nil {
			obj.onEvent.OnBlobFailed(w, r, outputPropObj, obj, nil)
			HandleError(w, r, outputPropObj, ErrorCodeRequestCheck, err.Error())
			return
		}
	}
	uu, err := obj.manager.MakeRequestUrl(ctx, dirName, fileName, kv, obj.privateSign, vs)
	//
	if err != nil {
		obj.onEvent.OnBlobFailed(w, r, outputPropObj, obj, nil)
		HandleError(w, r, outputPropObj, ErrorCodeMakeRequestUrl, "failed to make uploadurl")
	} else {
		outputPropObj.SetString("token", uu.String())
		w.Write(outputPropObj.ToJson())
		w.WriteHeader(http.StatusOK)
	}
}

func (obj *BlobHandler) HandleUploaded(w http.ResponseWriter, r *http.Request) {
	//
	miniPropObj := miniprop.NewMiniProp()
	res, e := obj.manager.CheckedCallback(r, obj.privateSign)
	if e != nil {
		obj.onEvent.OnBlobFailed(w, r, miniPropObj, obj, nil)
		HandleError(w, r, miniPropObj, ErrorCodeCheckCallback, e.Error())
		return
	}

	//
	ctx := appengine.NewContext(r)
	newItem := obj.manager.NewBlobItem(ctx, res.DirName, res.FileName, res.BlobKey)
	if obj.onEvent.OnBlobBeforeSave != nil {
		err := obj.onEvent.OnBlobBeforeSave(w, r, miniPropObj, obj, newItem)
		if err != nil {
			obj.onEvent.OnBlobFailed(w, r, miniPropObj, obj, newItem)
			HandleError(w, r, miniPropObj, ErrorCodeBeforeSaveCheck, "Failed to check")
			return
		}
	}
	err2 := obj.manager.SaveBlobItem(ctx, newItem)
	if err2 != nil {
		obj.onEvent.OnBlobFailed(w, r, miniPropObj, obj, newItem)
		HandleError(w, r, miniPropObj, ErrorCodeSaveBlobItem, "Failed to save blobitem")
		return
	}

	if obj.onEvent.OnBlobComplete != nil {
		err := obj.onEvent.OnBlobComplete(w, r, miniPropObj, obj, newItem)
		if err != nil {
			obj.onEvent.OnBlobFailed(w, r, miniPropObj, obj, newItem)
			HandleError(w, r, miniPropObj, ErrorCodeCompleteCheck, "Failed to save blobitem")
			return
		}
	}
	miniPropObj.SetString("blobkey", newItem.GetBlobKey())
	w.Write(miniPropObj.ToJson())
	w.WriteHeader(http.StatusOK)
}
