package msgp

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/codec"
	"github.com/vmihailenco/msgpack/v4"
)

type msgpCodec struct {
}

// 编码器的名称
func (m *msgpCodec) Name() string {
	return "msgp"
}

func (m *msgpCodec) MimeType() string {
	return "application/msgp"
}

// 将结构体编码为JSON的字节数组
func (m *msgpCodec) Encode(msgObj interface{}, ctx cellnet.ContextSet) (data interface{}, err error) {

	return msgpack.Marshal(msgObj)

}

// 将JSON的字节数组解码为结构体
func (m *msgpCodec) Decode(data interface{}, msgObj interface{}) error {

	return msgpack.Unmarshal(data.([]byte), msgObj)
}

func init() {

	// 注册编码器
	codec.RegisterCodec(new(msgpCodec))
}
