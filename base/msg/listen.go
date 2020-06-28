package msg

import (
	"fmt"
	"github.com/davyxu/cellnet"
	"github.com/greatwing/wing/base/log"
	"reflect"
	"sync"
)

var (
	listenerByID   sync.Map
	listenerCustom sync.Map
)

type CustomCallback func(param interface{})

//监听消息
func On(msgId int, callback cellnet.EventCallback) {
	var cbSlice []cellnet.EventCallback

	if data, ok := listenerByID.Load(msgId); ok {
		oldSlice := data.([]cellnet.EventCallback)
		cbSlice = make([]cellnet.EventCallback, len(oldSlice)+1)
		copy(cbSlice, oldSlice)
		cbSlice = append(cbSlice, callback)
	} else {
		cbSlice = []cellnet.EventCallback{callback}
	}

	listenerByID.Store(msgId, cbSlice)
}
func OnType(msgType reflect.Type, callback cellnet.EventCallback) {
	if msgType.Kind() == reflect.Ptr {
		msgType = msgType.Elem()
	}
	if msgType.Kind() != reflect.Struct {
		log.Error("the type of msg must be struct")
		return
	}

	meta := cellnet.MessageMetaByType(msgType)
	if meta == nil {
		panic(fmt.Sprintf("cannot find meta by type(%v), need register msg", msgType))
	}
	On(meta.ID, callback)
}

//停止监听消息
func Off(msgId int) {
	listenerByID.Delete(msgId)
}

func Process(ev cellnet.Event) {
	//查找绑定的msgId
	meta := cellnet.MessageMetaByMsg(ev.Message())
	if meta != nil {
		if data, ok := listenerByID.Load(meta.ID); ok {
			cbSlice := data.([]cellnet.EventCallback)
			for _, callback := range cbSlice {
				callback(ev)
			}
		}
	}
}

//监听自定义消息
func OnCustom(key interface{}, callback CustomCallback) {
	var cbSlice []CustomCallback

	if data, ok := listenerCustom.Load(key); ok {
		oldSlice := data.([]CustomCallback)
		cbSlice = make([]CustomCallback, len(oldSlice)+1)
		copy(cbSlice, oldSlice)
		cbSlice = append(cbSlice, callback)
	} else {
		cbSlice = []CustomCallback{callback}
	}

	listenerCustom.Store(key, cbSlice)
}

//停止监听自定义消息
func OffCustom(key interface{}) {
	listenerCustom.Delete(key)
}

//
func FireCustom(key, param interface{}) {
	if data, ok := listenerCustom.Load(key); ok {
		cbSlice := data.([]CustomCallback)
		for _, callback := range cbSlice {
			callback(param)
		}
	}
}
