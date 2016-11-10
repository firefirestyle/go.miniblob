package handler

import (
	//	"net/url"

	"io/ioutil"
	"net/http"

	"github.com/firefirestyle/go.miniprop"

	//	miniblob "github.com/firefirestyle/go.miniblob/blob"
	"google.golang.org/appengine"
	//	"google.golang.org/appengine/blobstore"
	"strconv"
	"time"
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
	//
	kv := strconv.FormatInt(time.Now().Unix(), 36)
	vs := map[string]string{}
	{
		vsTmp := map[string]string{}
		var err error = nil
		for _, f := range obj.onEvent.OnBlobRequestList {
			vsTmp, err = f(w, r, inputPropObj, outputPropObj, obj)
			if err != nil {
				for _, ff := range obj.onEvent.OnBlobFailedList {
					ff(w, r, outputPropObj, obj, nil)
				}
				HandleError(w, r, outputPropObj, ErrorCodeRequestCheck, err.Error())
				return
			}
			for k, v := range vsTmp {
				vs[k] = v
			}
		}
	}
	uu, err := obj.manager.MakeRequestUrl(ctx, dirName, fileName, kv, obj.privateSign, vs)
	//
	if err != nil {
		for _, ff := range obj.onEvent.OnBlobFailedList {
			ff(w, r, outputPropObj, obj, nil)
		}
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
		for _, ff := range obj.onEvent.OnBlobFailedList {
			ff(w, r, outputPropObj, obj, nil)
		}
		HandleError(w, r, outputPropObj, ErrorCodeCheckCallback, e.Error())
		return
	}
	curTime := time.Now().Unix()
	kvTime, errTime := strconv.ParseInt(r.FormValue("kv"), 36, 64)
	if errTime != nil || !(curTime-60*1 < kvTime && kvTime < curTime+60*10) {
		for _, ff := range obj.onEvent.OnBlobFailedList {
			ff(w, r, outputPropObj, obj, nil)
		}
		HandleError(w, r, outputPropObj, ErrorCodeCheckCallback, "kv time error")
		return
	}

	//
	ctx := appengine.NewContext(r)
	newItem := obj.manager.NewBlobItem(ctx, res.DirName, res.FileName, res.BlobKey)
	//
	if obj.onEvent.OnBlobBeforeSaveList != nil {
		for _, f := range obj.onEvent.OnBlobBeforeSaveList {
			err := f(w, r, outputPropObj, obj, newItem)
			if err != nil {
				for _, ff := range obj.onEvent.OnBlobFailedList {
					ff(w, r, outputPropObj, obj, newItem)
				}
				HandleError(w, r, outputPropObj, ErrorCodeBeforeSaveCheck, "Failed to check")
				return
			}
		}
	}
	err2 := obj.manager.SaveBlobItemWithImmutable(ctx, newItem)
	if err2 != nil {
		for _, ff := range obj.onEvent.OnBlobFailedList {
			ff(w, r, outputPropObj, obj, newItem)
		}
		HandleError(w, r, outputPropObj, ErrorCodeSaveBlobItem, "Failed to save blobitem")
		return
	}

	for _, f := range obj.onEvent.OnBlobCompleteList {
		err3 := f(w, r, outputPropObj, obj, newItem)
		if err3 != nil {
			for _, ff := range obj.onEvent.OnBlobFailedList {
				ff(w, r, outputPropObj, obj, newItem)
			}

			HandleError(w, r, outputPropObj, ErrorCodeCompleteCheck, "Failed to save blobitem")
			return
		}
	}
	outputPropObj.SetString("blobkey", newItem.GetBlobKey())
	w.Write(outputPropObj.ToJson())
	w.WriteHeader(http.StatusOK)
}
