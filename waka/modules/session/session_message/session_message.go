package session_message

import "github.com/golang/protobuf/proto"

type Close struct{}

type Send struct {
	Payload proto.Message
}

// ---------------------------------------------------------------------------------------------------------------------

type Closed struct{}

type Transport struct {
	Payload proto.Message
}

type FutureRequest struct {
	Payload proto.Message
	Respond func(proto.Message, error)
}
