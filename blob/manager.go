package miniblob

import (
	"golang.org/x/net/context"
	//	"google.golang.org/appengine"

	"errors"

	"github.com/firefirestyle/go.minipointer"
	"github.com/firefirestyle/go.miniprop"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type BlobManager struct {
	callbackUrl  string
	blobItemKind string
	rootGroup    string
	pointerMgr   *minipointer.PointerManager
}

type BlobManagerConfig struct {
	RootGroup     string
	Kind          string
	PointerKind   string
	CallbackUrl   string
	MemcachedOnly bool
}

func NewBlobManager(config BlobManagerConfig) *BlobManager {
	ret := new(BlobManager)
	ret.rootGroup = config.RootGroup
	ret.blobItemKind = config.Kind
	ret.callbackUrl = config.CallbackUrl
	ret.pointerMgr = minipointer.NewPointerManager(minipointer.PointerManagerConfig{
		RootGroup:     config.RootGroup,
		Kind:          config.PointerKind,
		MemcachedOnly: config.MemcachedOnly, // todo
	})
	return ret
}

func (obj *BlobManager) GetPointerMgr() *minipointer.PointerManager {
	return obj.pointerMgr
}

func (obj *BlobManager) GetBlobItem(ctx context.Context, parent string, name string, sign string) (*BlobItem, error) {
	key := obj.NewBlobItemGaeKey(ctx, parent, name, sign)

	return obj.GetBlobItemFromGaeKey(ctx, key)
}

func (obj *BlobManager) GetBlobItemFromQuery(ctx context.Context, parent string, name string) (*BlobItem, error) {
	founded := obj.FindBlobItemFromPath(ctx, parent, name, "")
	if len(founded.Keys) <= 0 {
		return nil, errors.New("not found blobitem")
	}
	key := obj.NewBlobItemGaeKeyFromStringId(ctx, founded.Keys[0])
	return obj.GetBlobItemFromGaeKey(ctx, key)
}

func (obj *BlobManager) GetBlobItemFromStringId(ctx context.Context, stringId string) (*BlobItem, error) {
	key := obj.NewBlobItemGaeKeyFromStringId(ctx, stringId)
	return obj.GetBlobItemFromGaeKey(ctx, key)
}

//
// if memcachedonly == true , posssible to become pointer == null
func (obj *BlobManager) GetBlobItemFromPointer(ctx context.Context, parent string, name string) (*BlobItem, *minipointer.Pointer, error) {
	pointerObj, pointerErr := obj.pointerMgr.GetPointer(ctx, obj.GetBlobId(parent, name), minipointer.TypePointer)
	if pointerErr != nil {
		if obj.pointerMgr.IsMemcachedOnly() == false {
			return nil, nil, pointerErr
		} else {
			o, e := obj.GetBlobItemFromQuery(ctx, parent, name)
			return o, nil, e
		}
	}
	retObj, retErr := obj.GetBlobItem(ctx, parent, name, pointerObj.GetSign())
	return retObj, pointerObj, retErr
}

func (obj *BlobManager) SaveBlobItemWithImmutable(ctx context.Context, newItem *BlobItem) error {
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
	errSave := newItem.saveDB(ctx)
	if errSave != nil {
		return errSave
	}

	//
	// pointer
	pointerObj := obj.pointerMgr.GetPointerForRelayId(ctx, obj.GetBlobId(newItem.GetParent(), newItem.GetName()))
	pointerObj.SetSign(newItem.GetBlobKey())
	pointerObj.SetValue(newItem.gaeObjectKey.StringID())
	pointerObj.SetOwner(newItem.gaeObject.Owner)
	pointerErr := obj.pointerMgr.Save(ctx, pointerObj)
	if pointerErr != nil {
		err := obj.DeleteBlobItemFromStringId(ctx, newItem.gaeObjectKey.StringID())
		if err != nil {
			Debug(ctx, "<gomidata>"+newItem.gaeObjectKey.StringID()+"</gomidata>")
		}
		return errSave
	}
	//
	// delete old data
	currItem, _, currErr := obj.GetBlobItemFromPointer(ctx, newItem.GetParent(), newItem.GetName())
	if currErr == nil {
		err := obj.DeleteBlobItem(ctx, currItem)
		if err != nil {
			Debug(ctx, "<gomidata>"+currItem.gaeObjectKey.StringID()+"</gomidata>")
		}
	}
	return nil

}

func (obj *BlobManager) GetPointer(ctx context.Context, parent, name string) (*minipointer.Pointer, error) {
	return obj.pointerMgr.GetPointer(ctx, obj.GetBlobId(parent, name), minipointer.TypePointer)
}

func (obj *BlobManager) GetPointerGaeKey(ctx context.Context, parent, name string) *datastore.Key {
	return obj.pointerMgr.NewPointerGaeKey(ctx, obj.GetBlobId(parent, name), minipointer.TypePointer)
}

func (obj *BlobManager) DeleteBlobItem(ctx context.Context, item *BlobItem) error {
	return obj.DeleteBlobItemFromStringId(ctx, item.gaeObjectKey.StringID())
}

func (obj *BlobManager) DeletePointer(ctx context.Context, parent, name string) error {
	return obj.GetPointerMgr().Delete(ctx, obj.GetBlobId(parent, name), minipointer.TypePointer)
}

func (obj *BlobManager) DeleteBlobItemWithPointer(ctx context.Context, item *BlobItem) error {
	return obj.DeleteBlobItemWithPointerFromStringId(ctx, item.gaeObjectKey.StringID())
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

func Debug(ctx context.Context, message string) {
	log.Infof(ctx, message)
}
