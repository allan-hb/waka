package tlv

import (
	"bytes"
	"encoding/binary"

	"github.com/davyxu/cellnet"
)

// Writer 实现了自定义封包的写入
type Writer struct{}

// Call 处理
func (w *Writer) Call(ev *cellnet.Event) {
	var outputHeadBuffer bytes.Buffer

	if err := binary.Write(&outputHeadBuffer, binary.LittleEndian, ev.MsgID); err != nil {
		ev.SetResult(cellnet.Result_PackageCrack)
		return
	}

	if err := binary.Write(&outputHeadBuffer, binary.LittleEndian, uint32(len(ev.Data))); err != nil {
		ev.SetResult(cellnet.Result_PackageCrack)
		return
	}

	binary.Write(&outputHeadBuffer, binary.LittleEndian, ev.Data)

	ev.Data = outputHeadBuffer.Bytes()
}

// NewWriter 创建一个封包写入者对象
func NewWriter() cellnet.EventHandler {
	return &Writer{}
}
