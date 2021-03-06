package blob

import (
	"golang.org/x/net/context"
	//	"google.golang.org/appengine"

	//	"errors"

	"github.com/firefirestyle/go.minipointer"
	"github.com/firefirestyle/go.miniprop"

	//"google.golang.org/appengine/datastore"
	//"google.golang.org/appengine/log"
)

func (obj *BlobManager) SavePointer(ctx context.Context, newItem *BlobItem) (*minipointer.Pointer, error) {
	pointerObj := obj.pointerMgr.GetPointerWithNewForRelayId(ctx, obj.MakeBlobId(newItem.GetParent(), newItem.GetName()))
	pointerObj.SetSign(newItem.GetBlobKey())
	pointerObj.SetValue(newItem.gaeKey.StringID())
	pointerObj.SetOwner(newItem.gaeObject.Owner)
	pointerErr := obj.pointerMgr.Save(ctx, pointerObj)
	return pointerObj, pointerErr
}

func (obj *BlobManager) SaveBlobItemWithImmutable(ctx context.Context, newItem *BlobItem) error {
	//
	// mkdirs
	pathObj := miniprop.NewMiniPath(newItem.GetParent())
	_, parentDirErr := obj.GetBlobItem(ctx, pathObj.GetDir(), ".dir", "")
	if parentDirErr != nil {
		for _, v := range pathObj.GetDirs() {
			dirObj := obj.NewBlobItem(ctx, v, ".dir", "")
			dirErr := dirObj.saveDB(ctx)
			if dirErr != nil {
				return dirErr
			}
		}
	}
	//
	//
	blobStringId, _, currErr := obj.GetBlobItemStringIdFromPointer(ctx, newItem.GetParent(), newItem.GetName())
	//currItem, _, currErr := obj.GetBlobItemFromPointer(ctx, newItem.GetParent(), newItem.GetName())

	errSave := newItem.saveDB(ctx)
	if errSave != nil {
		return errSave
	}

	//
	// pointer
	_, pointerErr := obj.SavePointer(ctx, newItem)
	if pointerErr != nil {
		err := obj.DeleteBlobItemFromStringId(ctx, newItem.gaeKey.StringID())
		if err != nil {
			Debug(ctx, "<gomidata>"+newItem.gaeKey.StringID()+"</gomidata>")
		}
		return errSave
	}
	//
	// delete old data

	if currErr == nil {
		err := obj.DeleteBlobItemFromStringId(ctx, blobStringId)
		if err != nil {
			Debug(ctx, "<gomidata>"+blobStringId+"</gomidata>")
		}
	}
	return nil

}

func (obj *BlobManager) DeleteBlobItem(ctx context.Context, item *BlobItem) error {
	return obj.DeleteBlobItemFromStringId(ctx, item.gaeKey.StringID())
}

func (obj *BlobManager) DeletePointer(ctx context.Context, parent, name string) error {
	return obj.GetPointerMgr().DeletePointer(ctx, obj.MakeBlobId(parent, name), minipointer.TypePointer)
}

func (obj *BlobManager) DeleteBlobItemWithPointer(ctx context.Context, item *BlobItem) error {
	return obj.DeleteBlobItemWithPointerFromStringId(ctx, item.gaeKey.StringID())
}

func (obj *BlobManager) DeleteBlobItemWithPointerFromStringId(ctx context.Context, stringId string) error {
	keyInfo := obj.GetKeyInfoFromStringId(stringId)
	obj.DeletePointer(ctx, keyInfo.Parent, keyInfo.Name)
	return obj.DeleteBlobItemFromStringId(ctx, stringId)
}

//
//
func (obj *BlobManager) DeleteBlobItemsWithPointerAtRecursiveMode(ctx context.Context, parent string) error {
	folders := make([]string, 0)
	folders = append(folders, parent)
	foldersTmp := make([]string, 0)
	for len(folders) > 0 {
		folder := folders[0]
		folders = folders[1:]
		foldersTmp = append(foldersTmp, folder)
		//
		founded := obj.FindAllBlobItemFromPath(ctx, folder)
		for _, v := range founded.Keys {
			keyInfo := obj.GetKeyInfoFromStringId(v)
			if keyInfo.Name == ".dir" {
				folders = append(folders, v)
				continue
			}
			blobObj, blobErr := obj.GetBlobItem(ctx, keyInfo.Parent, keyInfo.Name, keyInfo.Sign)
			if blobErr == nil {
				obj.DeleteBlobItemWithPointer(ctx, blobObj)
			}
		}
	}
	for _, v := range foldersTmp {
		keyInfo := obj.GetKeyInfoFromStringId(v)
		blobObj, blobErr := obj.GetBlobItem(ctx, keyInfo.Parent, keyInfo.Name, keyInfo.Sign)
		if blobErr == nil {
			obj.DeleteBlobItemWithPointer(ctx, blobObj)
		}
	}
	return nil
}
