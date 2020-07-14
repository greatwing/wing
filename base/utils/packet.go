package utils

import (
	"encoding/binary"
	"errors"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/codec"
	"github.com/davyxu/cellnet/util"
	"io"
)

var (
	ErrMaxPacket  = errors.New("packet over size")
	ErrMinPacket  = errors.New("packet short size")
	ErrShortMsgID = errors.New("short msgid")
)

const (
	bodySize  = 2 // 包体大小字段
	msgIDSize = 2 // 消息ID字段
)

func RecvLTVPacketData(reader io.Reader, maxPacketSize int) (msgID int, msgData []byte, err error) {

	// Size为uint16，占2字节
	var sizeBuffer = make([]byte, bodySize)

	// 持续读取Size直到读到为止
	_, err = io.ReadFull(reader, sizeBuffer)

	// 发生错误时返回
	if err != nil {
		return
	}

	if len(sizeBuffer) < bodySize {
		err = ErrMinPacket
		return
	}

	// 用小端格式读取Size
	size := binary.LittleEndian.Uint16(sizeBuffer)

	if maxPacketSize > 0 && size >= uint16(maxPacketSize) {
		err = ErrMaxPacket
		return
	}

	// 分配包体大小
	body := make([]byte, size) // TODO 内存池优化

	// 读取包体数据
	_, err = io.ReadFull(reader, body)

	// 发生错误时返回
	if err != nil {
		return
	}

	if len(body) < bodySize {
		err = ErrShortMsgID
		return
	}

	msgID = int(binary.LittleEndian.Uint16(body))

	msgData = body[msgIDSize:]

	return
}

// 接收Length-Type-Value格式的封包流程
func RecvLTVPacket(reader io.Reader, maxPacketSize int) (interface{}, error) {

	msgId, msgData, err := RecvLTVPacketData(reader, maxPacketSize)
	if err != nil {
		return nil, err
	}

	// 将字节数组和消息ID用户解出消息
	msg, _, err := codec.DecodeMessage(int(msgId), msgData)
	if err != nil {
		return nil, err
	}

	return msg, err
}

// 发送Length-Type-Value格式的封包流程
func SendLTVPacket(writer io.Writer, ctx cellnet.ContextSet, data interface{}) error {

	var (
		msgData []byte
		msgID   int
		meta    *cellnet.MessageMeta
	)

	switch m := data.(type) {
	case *cellnet.RawPacket: // 发裸包
		msgData = m.MsgData
		msgID = m.MsgID
	default: // 发普通编码包
		var err error

		// 将用户数据转换为字节数组和消息ID
		msgData, meta, err = codec.EncodeMessage(data, ctx)

		if err != nil {
			return err
		}

		msgID = meta.ID
	}

	pkt := make([]byte, bodySize+msgIDSize+len(msgData))

	// Length
	binary.LittleEndian.PutUint16(pkt, uint16(msgIDSize+len(msgData)))

	// Type
	binary.LittleEndian.PutUint16(pkt[bodySize:], uint16(msgID))

	// Value
	copy(pkt[bodySize+msgIDSize:], msgData)

	// 将数据写入Socket
	err := util.WriteFull(writer, pkt)

	// Codec中使用内存池时的释放位置
	if meta != nil {
		codec.FreeCodecResource(meta.Codec, msgData, ctx)
	}

	return err
}
