package miniblob

import (
	"strings"

	"net/url"

	"encoding/base64"

	"golang.org/x/net/context"

	"net/http"

	"errors"

	//	"google.golang.org/appengine"
	"google.golang.org/appengine/blobstore"
)

//
//
//
func (obj *BlobManager) MakeRequestUrl(ctx context.Context, dirName string, fileName string, opt string) (string, error) {
	if opt == "" {
		opt = "none"
	}

	var ary = []string{obj.BasePath + //
		"?dir=", url.QueryEscape(base64.StdEncoding.EncodeToString([]byte(dirName))), //
		"&file=", url.QueryEscape(fileName), //
		"&opt=", opt}
	uu, err2 := blobstore.UploadURL(ctx, strings.Join(ary, ""), nil) //&option)
	return uu.String(), err2
}

//
//
//
func (obj *BlobManager) HandleUploaded(ctx context.Context, r *http.Request) (*BlobItem, string, error) {
	//
	blobs, _, err := blobstore.ParseUpload(r)
	if err != nil {
		return nil, "", err
	}

	// --
	// dirName
	// --
	dirNameSrc, err1 := base64.StdEncoding.DecodeString(r.FormValue("dir"))
	if err1 != nil {
		return nil, "", err1
	}
	dirName := string(dirNameSrc)

	// --
	// filename
	// --
	fileName := r.FormValue("file")

	// --
	// opt
	// --
	optProp := string(r.FormValue("opt"))

	// --
	// file
	// --
	file := blobs["file"]
	if len(file) == 0 {
		return nil, "", errors.New("")
	}
	blobKey := string(file[0].BlobKey)
	if fileName == "" {
		fileName = blobKey
	}

	//
	//
	//
	newItem := obj.NewBlobItem(ctx, dirName, fileName, blobKey)
	err2 := obj.SaveBlobItem(ctx, newItem)
	if err2 != nil {
		return nil, "", errors.New("")
	}
	return newItem, optProp, err
}
