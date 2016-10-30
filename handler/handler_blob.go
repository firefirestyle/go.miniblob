package handler

import (
	//	"net/url"

	"io/ioutil"
	"net/http"

	"github.com/firefirestyle/go.miniprop"

	//	miniblob "github.com/firefirestyle/go.miniblob/blob"
	"google.golang.org/appengine"
	//	"google.golang.org/appengine/blobstore"
)

func (obj *BlobHandler) HandleBlobRequestToken(w http.ResponseWriter, r *http.Request) {
	params, _ := ioutil.ReadAll(r.Body)
	inputPropObj := miniprop.NewMiniPropFromJson(params)
	dirName := inputPropObj.GetString("dir", "")
	fileName := inputPropObj.GetString("file", "")
	obj.HandleBlobRequestTokenFromParams(w, r, dirName, fileName, inputPropObj)
}

func (obj *BlobHandler) HandleBlobRequestTokenFromParams(w http.ResponseWriter, r *http.Request, dirName string, fileName string, inputPropObj *miniprop.MiniProp) {
	ctx := appengine.NewContext(r)
	outputPropObj := miniprop.NewMiniProp()
	if inputPropObj == nil {
		params, _ := ioutil.ReadAll(r.Body)
		inputPropObj = miniprop.NewMiniPropFromJson(params)
	}
	//
	kv := "abcdef"
	vs := map[string]string{}
	{
		var err error = nil
		kv, vs, err = obj.onEvent.OnBlobRequest(w, r, inputPropObj, outputPropObj, obj)
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
	outputPropObj := miniprop.NewMiniProp()
	res, e := obj.manager.CheckedCallback(r, obj.privateSign)
	if e != nil {
		obj.onEvent.OnBlobFailed(w, r, outputPropObj, obj, nil)
		HandleError(w, r, outputPropObj, ErrorCodeCheckCallback, e.Error())
		return
	}

	//
	ctx := appengine.NewContext(r)
	newItem := obj.manager.NewBlobItem(ctx, res.DirName, res.FileName, res.BlobKey)
	//
	if obj.onEvent.OnBlobBeforeSave != nil {
		err := obj.onEvent.OnBlobBeforeSave(w, r, outputPropObj, obj, newItem)
		if err != nil {
			obj.onEvent.OnBlobFailed(w, r, outputPropObj, obj, newItem)
			HandleError(w, r, outputPropObj, ErrorCodeBeforeSaveCheck, "Failed to check")
			return
		}
	}
	err2 := obj.manager.SaveBlobItemWithImmutable(ctx, newItem)
	if err2 != nil {
		obj.onEvent.OnBlobFailed(w, r, outputPropObj, obj, newItem)
		HandleError(w, r, outputPropObj, ErrorCodeSaveBlobItem, "Failed to save blobitem")
		return
	}

	Debug(ctx, "onBlobComplete --s")
	err3 := obj.onEvent.OnBlobComplete(w, r, outputPropObj, obj, newItem)
	if err3 != nil {
		obj.onEvent.OnBlobFailed(w, r, outputPropObj, obj, newItem)
		HandleError(w, r, outputPropObj, ErrorCodeCompleteCheck, "Failed to save blobitem")
		return
	}
	outputPropObj.SetString("blobkey", newItem.GetBlobKey())
	w.Write(outputPropObj.ToJson())
	w.WriteHeader(http.StatusOK)
}
