package gateway_message

// 连接已关闭
type Closed struct{}

// 心跳
type Heart struct{}

// 传输
type Transport struct {
	Id      uint32
	Payload []byte
}

// RPC
type FutureRequest struct {
	Id      uint32
	Payload []byte
	Number  uint64
}
