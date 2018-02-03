package session

import (
	"errors"
	"fmt"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka/codec"
	"github.com/liuhan907/waka/waka/modules/gateway/gateway_message"
	"github.com/liuhan907/waka/waka/modules/session/session_message"
	"github.com/liuhan907/waka/waka/proto"
)

var (
	errResponseIsNil = errors.New("response is nil")
)

func (my *actorT) ReceiveGateway(context actor.Context) bool {
	switch ev := context.Message().(type) {
	case *gateway_message.Heart:
		my.heartbeat()
	case *gateway_message.Closed:
		my.closed()
	case *gateway_message.Transport:
		my.transport(ev)
	case *gateway_message.FutureRequest:
		my.futureRequest(ev)
	default:
		return false
	}
	return true
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) heartbeat() {
	if my.option.EnableHeartLog {
		log.Debugln("heartbeat")
	}

	my.heart = time.Now()
}

func (my *actorT) closed() {
	if my.option.EnableLog {
		log.WithFields(logrus.Fields{
			"pid": my.pid.String(),
		}).Debugln("session closed")
	}
	my.pid.Stop()

	my.target.Tell(&session_message.Closed{})
}

func (my *actorT) transport(ev *gateway_message.Transport) {
	m, name, err := codec.Decode(ev.Id, ev.Payload)
	if err != nil {
		if my.option.EnableLog {
			log.WithFields(logrus.Fields{
				"id":      ev.Id,
				"payload": ev.Payload,
				"err":     err,
			}).Warnln("decode transport failed")
		}
		return
	}

	if my.option.EnableLog {
		log.WithFields(logrus.Fields{
			"id":      ev.Id,
			"name":    name,
			"payload": m.String(),
		}).Debugln("redirect transport from gateway to target")
	}

	my.target.Tell(&session_message.Transport{m})
}

func (my *actorT) futureRequest(ev *gateway_message.FutureRequest) {
	m, name, err := codec.Decode(ev.Id, ev.Payload)
	if err != nil {
		if my.option.EnableLog {
			log.WithFields(logrus.Fields{
				"id":      ev.Id,
				"payload": ev.Payload,
				"number":  ev.Number,
				"err":     err,
			}).Warnln("decode future request failed")
		}
		return
	}

	if my.option.EnableLog {
		log.WithFields(logrus.Fields{
			"id":      ev.Id,
			"name":    name,
			"payload": m.String(),
			"number":  ev.Number,
		}).Debugln("redirect future request from gateway to target")
	}

	ch := make(chan interface{})
	respond := func(m proto.Message, e error) {
		if e != nil {
			ch <- e
		} else if m != nil {
			ch <- m
		} else {
			ch <- errResponseIsNil
		}
		close(ch)
	}

	my.target.Tell(&session_message.FutureRequest{m, respond})

	response := <-ch

	if err, ok := response.(error); ok {
		if my.option.EnableLog {
			log.WithFields(logrus.Fields{
				"number": ev.Number,
				"err":    err,
			}).Warnln("future response failed")
		}

		my.conn.Send(&waka_proto.FutureResponse{
			Status: fmt.Sprintf("failed: %v", err),
			Number: ev.Number,
		})
	} else if m, ok := response.(proto.Message); ok {
		d, id, name, err := codec.Encode(m)
		if err != nil {
			if my.option.EnableLog {
				log.WithFields(logrus.Fields{
					"number":  ev.Number,
					"payload": m.String(),
					"err":     err,
				}).Warnln("future response encode failed")
			}

			my.conn.Send(&waka_proto.FutureResponse{
				Status: fmt.Sprintf("failed: encode failed: %v", err),
				Number: ev.Number,
			})
		} else {
			if my.option.EnableLog {
				log.WithFields(logrus.Fields{
					"id":      id,
					"name":    name,
					"payload": m.String(),
					"number":  ev.Number,
				}).Debugln("redirect future response from target to gateway")
			}

			my.conn.Send(&waka_proto.FutureResponse{
				Status:  "success",
				Id:      id,
				Payload: d,
				Number:  ev.Number,
			})
		}
	} else {
		if my.option.EnableLog {
			log.WithFields(logrus.Fields{
				"number": ev.Number,
			}).Warnln("future response illegal")
		}

		my.conn.Send(&waka_proto.FutureResponse{
			Status: fmt.Sprintf("failed: response illegal"),
			Number: ev.Number,
		})
	}
}
