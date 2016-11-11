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
	vs := map[string]string{}
	reqCheckRet, reqCheckErr := obj.OnBlobRequestList(w, r, inputPropObj, outputPropObj, obj)
	for k, v := range reqCheckRet {
		vs[k] = v
	}

	if reqCheckErr != nil {
		obj.OnBlobFailed(w, r, outputPropObj, obj, nil)
		HandleError(w, r, outputPropObj, ErrorCodeRequestCheck, reqCheckErr.Error())
		return
	}

	//
	//
	kv := strconv.FormatInt(time.Now().Unix(), 36)
	reqUrl, reqName, err := obj.manager.MakeRequestUrl(ctx, dirName, fileName, kv, obj.privateSign, vs)
	if err != nil {
		obj.OnBlobFailed(w, r, outputPropObj, obj, nil)
		HandleError(w, r, outputPropObj, ErrorCodeMakeRequestUrl, "failed to make uploadurl")
	} else {
		outputPropObj.SetString("token", reqUrl.String())
		outputPropObj.SetString("name", reqName)
		w.Write(outputPropObj.ToJson())
		w.WriteHeader(http.StatusOK)
	}
}

func (obj *BlobHandler) HandleUploaded(w http.ResponseWriter, r *http.Request) {
	//
	//
	outputPropObj := miniprop.NewMiniProp()
	res, e := obj.manager.CheckedCallback(r, obj.privateSign)
	if e != nil {
		obj.OnBlobFailed(w, r, outputPropObj, obj, nil)
		HandleError(w, r, outputPropObj, ErrorCodeCheckCallback, e.Error())
		return
	}
	curTime := time.Now().Unix()
	kvTime, errTime := strconv.ParseInt(r.FormValue("kv"), 36, 64)
	if errTime != nil || !(curTime-60*1 < kvTime && kvTime < curTime+60*10) {
		obj.OnBlobFailed(w, r, outputPropObj, obj, nil)
		HandleError(w, r, outputPropObj, ErrorCodeCheckCallback, "kv time error")
		return
	}

	//
	ctx := appengine.NewContext(r)
	newItem := obj.manager.NewBlobItem(ctx, res.DirName, res.FileName, res.BlobKey)
	//
	befErr := obj.OnBlobBeforeSave(w, r, outputPropObj, obj, newItem)
	if befErr != nil {
		obj.OnBlobFailed(w, r, outputPropObj, obj, newItem)
		HandleError(w, r, outputPropObj, ErrorCodeBeforeSaveCheck, befErr.Error())
		return
	}
	err2 := obj.manager.SaveBlobItemWithImmutable(ctx, newItem)
	if err2 != nil {
		obj.OnBlobFailed(w, r, outputPropObj, obj, newItem)
		HandleError(w, r, outputPropObj, ErrorCodeSaveBlobItem, err2.Error())
		return
	}

	err3 := obj.OnBlobComplete(w, r, outputPropObj, obj, newItem)
	if err3 != nil {
		obj.OnBlobFailed(w, r, outputPropObj, obj, newItem)
		HandleError(w, r, outputPropObj, ErrorCodeCompleteCheck, err3.Error())
		return
	}
	outputPropObj.SetString("blobkey", newItem.GetBlobKey())
	w.Write(outputPropObj.ToJson())
	w.WriteHeader(http.StatusOK)
}
