package service

import (
	"errors"
	"fmt"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/codec"
	_ "github.com/greatwing/wing/base/codec/msgp"
	"reflect"
)

type ServiceIdentifyACK struct {
	SvcName string
	SvcID   string
}

func (self *ServiceIdentifyACK) String() string { return fmt.Sprintf("%+v", *self) }

func init() {
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: codec.MustGetCodec("msgp"),
		Type:  reflect.TypeOf((*ServiceIdentifyACK)(nil)).Elem(),
		ID:    1000, //todo
	})
}

var (
	ErrInvalidRelayMessage         = errors.New("invalid relay message")
	ErrInvalidRelayPassthroughType = errors.New("invalid relay passthrough type")
)