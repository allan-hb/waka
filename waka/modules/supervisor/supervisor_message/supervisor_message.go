package supervisor_message

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
)

// 玩家通知监督者玩家进入
type PlayerEnter struct {
	Conn   *actor.PID
	Player uint64
	Remote string
}

// 玩家通知监督者玩家离开
type PlayerLeave struct {
	Player uint64
}

// 玩家通知监督者数据传输
type PlayerTransport struct {
	Player  uint64
	Payload proto.Message
}

// 玩家通知监督者 RPC 请求传输
type PlayerFutureRequest struct {
	Player  uint64
	Payload proto.Message
	Respond func(proto.Message, error)
}

// ---------------------------------------------------------------------------------------------------------------------

// 监督者通知玩家关闭会话
type Close struct{}

// 监督者通知玩家传输数据
type SendFromSupervisor struct {
	Payload proto.Message
}

// ---------------------------------------------------------------------------------------------------------------------

// 监督者通知大厅玩家进入
type PlayerEntered struct {
	Player uint64
	Remote string
}

// 监督者通知大厅玩家变更
type PlayerExchanged struct {
	Player uint64
	Remote string
}

// 监督者通知大厅玩家离开
type PlayerLeft struct {
	Player uint64
}

// 监督者通知大厅数据传输
type PlayerTransported struct {
	Player  uint64
	Payload proto.Message
}

// 监督者通知大厅 RPC 数据传输
type PlayerFutureRequested struct {
	Player  uint64
	Payload proto.Message
	Respond func(proto.Message, error)
}

// ---------------------------------------------------------------------------------------------------------------------

// 大厅通知监督者向玩家发送数据
type SendFromHall struct {
	Player  uint64
	Payload proto.Message
}
