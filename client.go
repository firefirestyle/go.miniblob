package miniblob

import (
	"bytes"
	"net/http"
	//"net/url"

	"golang.org/x/net/context"

	"mime/multipart"

	//"io"

	"google.golang.org/appengine/blobstore"
	"google.golang.org/appengine/urlfetch"
)

//func (obj *BlobManager) MakeRequestUrlForOwn(ctx context.Context, dirName string, fileName string, //
//	publicSign string, privateSign string, optKeyValues map[string]string) (*url.URL, error) {
//	return obj.MakeRequestUrl(ctx, dirName, fileName, publicSign, privateSign, optKeyValues)
//}

func (obj *BlobManager) MakeRequestUrlForOwn(ctx context.Context, dirName string, fileName string, data []byte) error {
	urlObj, err := blobstore.UploadURL(ctx, "/dummy", nil)
	if err != nil {
		return err
	}
	return obj.SaveData(ctx, urlObj.String(), data)
}

func (obj *BlobManager) SaveData(c context.Context, url string, data []byte) error {

	var b bytes.Buffer
	fw := multipart.NewWriter(&b)
	file, err := fw.CreateFormField("file")
	if err != nil {
		return err
	}
	if _, err = file.Write(data); err != nil {
		return err
	}
	fw.Close()
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", fw.FormDataContentType())

	//
	//
	client := urlfetch.Client(c)
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusCreated {
		return err
	}
	return nil
}
