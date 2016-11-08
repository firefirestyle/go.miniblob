package miniblob

import (
	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
)

type BlobFounds struct {
	Keys       []string
	CursorNext string
	CursorOne  string
}

/*
https://cloud.google.com/appengine/docs/go/config/indexconfig#updating_indexes
*/
func (obj *BlobManager) FindBlobItemFromParent(ctx context.Context, parent string, cursorSrc string) BlobFounds {
	//
	q := datastore.NewQuery(obj.blobItemKind)
	q = q.Filter("ProjectId =", obj.rootGroup)
	q = q.Filter("Parent =", parent)
	q = q.Order("-Updated")
	//
	return obj.FindBlobItemFromQuery(ctx, q, cursorSrc)
}

func (obj *BlobManager) FindBlobItemFromPath(ctx context.Context, parent string, name string, cursorSrc string) BlobFounds {
	//
	q := datastore.NewQuery(obj.blobItemKind)
	q = q.Filter("ProjectId =", obj.rootGroup)
	q = q.Filter("Parent =", parent)
	q = q.Filter("Name =", name)
	q = q.Order("-Updated")
	//
	return obj.FindBlobItemFromQuery(ctx, q, cursorSrc)
}

//
//
func (obj *BlobManager) FindBlobItemFromQuery(ctx context.Context, q *datastore.Query, cursorSrc string) BlobFounds {
	cursor := obj.newCursorFromSrc(cursorSrc)
	if cursor != nil {
		q = q.Start(*cursor)
	}
	q = q.KeysOnly()
	founds := q.Run(ctx)

	var keys []string
	var cursorNext string = ""
	var cursorOne string = ""

	for i := 0; ; i++ {
		key, err := founds.Next(nil)
		if err != nil || err == datastore.Done {
			break
		} else {
			keys = append(keys, key.StringID())
		}
		if i == 0 {
			cursorOne = obj.makeCursorSrc(founds)
		}
	}
	cursorNext = obj.makeCursorSrc(founds)
	return BlobFounds{
		Keys:       keys,
		CursorOne:  cursorOne,
		CursorNext: cursorNext,
	}
}

func (obj *BlobManager) newCursorFromSrc(cursorSrc string) *datastore.Cursor {
	c1, e := datastore.DecodeCursor(cursorSrc)
	if e != nil {
		return nil
	} else {
		return &c1
	}
}

func (obj *BlobManager) makeCursorSrc(founds *datastore.Iterator) string {
	c, e := founds.Cursor()
	if e == nil {
		return c.String()
	} else {
		return ""
	}
}
