package tlv

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/socket"
)

// Reader 实现了自定义封包的读取
type Reader struct{}

// Call 处理
func (r *Reader) Call(ev *cellnet.Event) {

	headReader := bytes.NewReader(ev.Data)

	if err := binary.Read(headReader, binary.LittleEndian, &ev.MsgID); err != nil {
		ev.SetResult(cellnet.Result_PackageCrack)
		return
	}

	var bodySize uint32
	if err := binary.Read(headReader, binary.LittleEndian, &bodySize); err != nil {
		ev.SetResult(cellnet.Result_PackageCrack)
		return
	}

	maxPacketSize := ev.Ses.FromPeer().(socket.SocketOptions).MaxPacketSize()
	if maxPacketSize > 0 && int(bodySize) > maxPacketSize {
		ev.SetResult(cellnet.Result_PackageCrack)
		return
	}

	reader := ev.Ses.(interface {
		DataSource() io.ReadWriter
	}).DataSource()

	dataBuffer := make([]byte, bodySize)
	if _, err := io.ReadFull(reader, dataBuffer); err != nil {
		ev.SetResult(cellnet.Result_PackageCrack)
		return
	}

	ev.Data = dataBuffer
}

// NewReader 创建一个封包读取者对象
func NewReader() cellnet.EventHandler {
	return &Reader{}
}
