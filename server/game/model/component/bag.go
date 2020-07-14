package component

import (
	"github.com/greatwing/wing/proto"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BagComponent struct {
	Items   []*proto.ItemData          `bson:"items,omitempty"`
	itemMap map[string]*proto.ItemData `bson:"-"`

	//updateItems map[string]struct{}
	//deleteItems map[string]struct{}
}

func (b *BagComponent) RebuildItemMap() {
	b.itemMap = make(map[string]*proto.ItemData)
	for _, item := range b.Items {
		b.itemMap[item.Id] = item
	}

	//b.updateItems = make(map[string]struct{})
	//b.deleteItems = make(map[string]struct{})
}

func (b *BagComponent) AddItem(itemData *proto.ItemData) {
	if _, ok := b.itemMap[itemData.Id]; ok {
		//这个道具已在背包里
		return
	}

	b.Items = append(b.Items, itemData)
	b.itemMap[itemData.Id] = itemData
	//b.updateItems[itemData.UUID] = struct{}{}
}

func (b *BagComponent) NewItem(itemStaticID, count int32) *proto.ItemData {
	item := &proto.ItemData{
		Id:       primitive.NewObjectID().Hex(),
		StaticId: itemStaticID,
		Count:    count,
	}
	b.AddItem(item)
	return item
}

func (b *BagComponent) RemoveItem(uuid string) {
	delete(b.itemMap, uuid)
	for index, data := range b.Items {
		if data.Id == uuid {
			b.Items = append(b.Items[:index], b.Items[index+1:]...)
			break
		}
	}
	//b.deleteItems[uuid] = struct{}{}
}

func (b *BagComponent) UpdateItem(itemData *proto.ItemData) {
	if _, ok := b.itemMap[itemData.Id]; !ok {
		//背包里没这个道具
		return
	}

	//b.updateItems[itemData.UUID] = struct{}{}
}

func (b *BagComponent) GetItem(uuid string) *proto.ItemData {
	return b.itemMap[uuid]
}

//func (b *BagComponent) AppendItems(items []*proto.ItemData) {
//	b.Items = append(b.Items, items...)
//	for _, itemData := range items {
//		b.itemMap[itemData.UUID] = itemData
//	}
//}

//func (b *BagComponent) GetDeletedItem() []string {
//	result := make([]string, 0, len(b.deleteItems))
//	for key, _ := range b.deleteItems {
//		result = append(result, key)
//	}
//	return result
//}
//
//func (b *BagComponent) GetUpdatedItem() []*proto.ItemData {
//	result := make([]*proto.ItemData, 0, len(b.updateItems))
//	for key, _ := range b.updateItems {
//		if item := b.GetItem(key); item != nil {
//			result = append(result, item)
//		}
//	}
//	return result
//}
