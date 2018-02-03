package codec

import (
	"reflect"

	"github.com/davyxu/cellnet"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

var (
	ErrNotIllegalMessage = errors.New("antares: decoded message type not proto.Message")
)

func Encode(m proto.Message) ([]byte, uint32, string, error) {
	meta := cellnet.MessageMetaByType(reflect.TypeOf(m))
	if meta == nil {
		return nil, 0, "", cellnet.ErrMessageNotFound
	}
	if meta.Codec == nil {
		return nil, 0, "", cellnet.ErrCodecNotFound
	}

	data, err := meta.Codec.Encode(m)
	if err != nil {
		return nil, 0, "", err
	}
	return data, meta.ID, meta.Name, nil
}

func Decode(id uint32, data []byte) (proto.Message, string, error) {
	meta := cellnet.MessageMetaByID(id)
	if meta == nil {
		return nil, "", cellnet.ErrMessageNotFound
	}
	if meta.Codec == nil {
		return nil, "", cellnet.ErrCodecNotFound
	}

	payload := reflect.New(meta.Type).Interface()
	err := meta.Codec.Decode(data, payload)
	if err != nil {
		return nil, "", err
	}

	m, ok := payload.(proto.Message)
	if !ok {
		return nil, "", ErrNotIllegalMessage
	}

	return m, meta.Name, nil
}
