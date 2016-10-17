package miniblob

import (
	"golang.org/x/net/context"

	"bytes"
	"net/http"

	"mime/multipart"

	"google.golang.org/appengine/urlfetch"
)

func (obj *BlobManager) SaveData(c context.Context, url string, sampleData []byte) error {

	// Now you can prepare a form that you will submit to that URL.
	var b bytes.Buffer
	fw := multipart.NewWriter(&b)
	// Do not change the form field, it must be "file"!
	// You are free to change the filename though, it will be stored in the BlobInfo.
	file, err := fw.CreateFormFile("file", "example.csv")
	if err != nil {
		return err
	}
	if _, err = file.Write(sampleData); err != nil {
		return err
	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	fw.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return err
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", fw.FormDataContentType())

	// Now submit the request.
	client := urlfetch.Client(c)
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	// Check the response status, it should be whatever you return in the `/upload` handler.
	if res.StatusCode != http.StatusCreated {
		return err
	}
	// Everything went fine.
	return nil
}
