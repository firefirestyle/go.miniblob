package miniblob

import (
	"golang.org/x/net/context"
	//	"google.golang.org/appengine"

	"github.com/firefirestyle/go.minipointer"
	//	"github.com/firefirestyle/go.miniprop"
	"google.golang.org/appengine/log"
)

type BlobManager struct {
	callbackUrl  string
	blobItemKind string
	rootGroup    string
	pointerMgr   *minipointer.PointerManager
}

type BlobManagerConfig struct {
	RootGroup   string
	Kind        string
	PointerKind string
	CallbackUrl string
}

func NewBlobManager(config BlobManagerConfig) *BlobManager {
	ret := new(BlobManager)
	ret.rootGroup = config.RootGroup
	ret.blobItemKind = config.Kind
	ret.callbackUrl = config.CallbackUrl
	ret.pointerMgr = minipointer.NewPointerManager(minipointer.PointerManagerConfig{
		RootGroup: config.RootGroup,
		Kind:      config.PointerKind,
		//MemcachedOnly: true, // todo
	})
	return ret
}

func (obj *BlobManager) GetPointerMgr() *minipointer.PointerManager {
	return obj.pointerMgr
}

func (obj *BlobManager) GetBlobItem(ctx context.Context, parent string, name string, sign string) (*BlobItem, error) {
	key := obj.NewBlobItemKey(ctx, parent, name, sign)

	return obj.NewBlobItemFromGaeObjectKey(ctx, key)
}

/*
func (obj *BlobManager) GetBlobItemFromQuery(ctx context.Context, parent string, name string) (*BlobItem, error) {
	obj.FindBlobItemFromPath(ctx, parent, name, "")
	key := obj.NewBlobItemKey(ctx, parent, name, sign)

	return obj.NewBlobItemFromGaeObjectKey(ctx, key)
}
*/
func (obj *BlobManager) GetBlobItemFromPointer(ctx context.Context, parent string, name string) (*BlobItem, *minipointer.Pointer, error) {
	pointerObj, pointerErr := obj.pointerMgr.GetPointer(ctx, obj.GetBlobId(parent, name), minipointer.TypePointer)
	if pointerErr != nil {
		//		if obj.pointerMgr.IsMemcachedOnly() == false {
		return nil, nil, pointerErr
		//		} else {
		//			obj.GetBlobItemFromP
		//		}
	}

	retObj, retErr := obj.GetBlobItem(ctx, parent, name, pointerObj.GetSign())
	return retObj, pointerObj, retErr
}

func (obj *BlobManager) SaveBlobItemWithImmutable(ctx context.Context, newItem *BlobItem) error {
	errSave := newItem.saveDB(ctx)
	if errSave != nil {
		return errSave
	}

	currItem, _, currErr := obj.GetBlobItemFromPointer(ctx, newItem.GetParent(), newItem.GetName())
	pointerObj := obj.pointerMgr.GetPointerForRelayId(ctx, obj.GetBlobId(newItem.GetParent(), newItem.GetName()))

	pointerObj.SetSign(newItem.GetBlobKey())
	pointerObj.SetValue(newItem.gaeObjectKey.StringID())
	pointerObj.SetOwner(newItem.gaeObject.Owner)
	pointerErr := obj.pointerMgr.Save(ctx, pointerObj)
	if pointerErr != nil {
		err := newItem.deleteFromDB(ctx)
		if err != nil {
			Debug(ctx, "<gomidata>"+newItem.gaeObjectKey.StringID()+"</gomidata>")
		}
		return errSave
	}

	if currErr != nil {
		Debug(ctx, "===> SIGN A")
		return nil
	} else {
		Debug(ctx, "===> SIGN B")
		err := obj.DeleteBlobItem(ctx, currItem)
		if err != nil {
			Debug(ctx, "<gomidata>"+currItem.gaeObjectKey.StringID()+"</gomidata>")
		}
		return nil
	}
}

func (obj *BlobManager) DeleteBlobItem(ctx context.Context, item *BlobItem) error {
	return item.deleteFromDB(ctx)
}

//
//

//
// you must to delete file before call this method, if there are many articleid's file.
//
func (obj *BlobManager) DeleteBlobItemsFormOnwer(ctx context.Context, owner string) error {
	pointerMgr := obj.GetPointerMgr()
	cursor := ""
	founded := pointerMgr.FindFromOwner(ctx, cursor, owner)
	for {
		if len(founded.Keys) <= 0 {
			Debug(ctx, "<E1>")
			break
		}
		for _, v := range founded.Keys {
			Debug(ctx, "<K> "+v)
			pointerKeyInfo := pointerMgr.GetKeyInfoFromStringId(v)
			blobitemKeyInfo := obj.GetKeyInfoFromStringId(pointerKeyInfo.Identify)
			//
			blobitemObj, pointerObj, _ := obj.GetBlobItemFromPointer(ctx, blobitemKeyInfo.Parent, blobitemKeyInfo.Name)
			if blobitemObj != nil {
				obj.DeleteBlobItem(ctx, blobitemObj)
			}
			if pointerObj != nil {
				pointerMgr.Delete(ctx, pointerKeyInfo.Identify, pointerKeyInfo.IdentifyType)
			}
		}
		prevFounded := founded
		founded = pointerMgr.FindFromOwner(ctx, founded.CursorNext, owner)
		if founded.CursorOne == prevFounded.CursorOne {
			Debug(ctx, "<E2>")
			break
		}
	}
	return nil
}

func Debug(ctx context.Context, message string) {
	log.Infof(ctx, message)
}
