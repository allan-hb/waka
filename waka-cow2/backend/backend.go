package backend

import (
	"net"
	"net/http"
	"os"
	"reflect"
	"strconv"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka-cow2/database"
	"github.com/liuhan907/waka/waka-cow2/modules/hall/hall_message"
)

var (
	log = logrus.WithFields(logrus.Fields{
		"pid":    os.Getpid(),
		"module": "cow2.backend",
	})
)

type httpHandler struct {
	target *actor.PID
}

func (w *httpHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	err := request.ParseForm()
	if err != nil {
		log.WithFields(logrus.Fields{
			"method":      request.Method,
			"request_uri": request.RequestURI,
			"err":         err,
		}).Warnln("form parse failed")
	}

	log.WithFields(logrus.Fields{
		"method":      request.Method,
		"request_uri": request.RequestURI,
		"path":        request.URL.Path,
		"form":        request.Form,
	}).Debugln("http request")

	switch request.Method {
	case "GET":
		switch request.URL.Path {
		case "/getTotalOnlineNums":
			w.getTotalOnlineNums(response, request)
		case "/getTotalOnlineId":
			w.getTotalOnlineId(response, request)
		case "/getTotalRoomInfo":
			w.getTotalRoomInfo(response, request)
		case "/playerChanged":
			w.playerChanged(response, request)
		default:
			response.WriteHeader(405)
		}
	default:
		response.WriteHeader(405)
	}
}

func (w *httpHandler) getTotalOnlineNums(response http.ResponseWriter, request *http.Request) {
	ch := make(chan interface{})
	w.target.Tell(&hall_message.GetTotalOnlineNumber{
		Respond: func(payload string, err error) {
			if err != nil {
				ch <- err
			} else {
				ch <- payload
			}
			close(ch)
		},
	})
	respond := <-ch
	switch evd := respond.(type) {
	case error:
		response.WriteHeader(400)

		log.WithFields(logrus.Fields{
			"method":      request.Method,
			"request_uri": request.RequestURI,
			"form":        request.Form,
			"err":         evd,
		}).Warnln("respond failed")
	case string:
		response.Write([]byte(evd))
	default:
		response.WriteHeader(400)

		log.WithFields(logrus.Fields{
			"method":      request.Method,
			"request_uri": request.RequestURI,
			"form":        request.Form,
			"respond":     reflect.TypeOf(respond),
			"err":         evd,
		}).Warnln("respond unknown")
	}
}

func (w *httpHandler) getTotalOnlineId(response http.ResponseWriter, request *http.Request) {
	ch := make(chan interface{})
	w.target.Tell(&hall_message.GetTotalOnline{
		Respond: func(payload string, err error) {
			if err != nil {
				ch <- err
			} else {
				ch <- payload
			}
			close(ch)
		},
	})
	respond := <-ch
	switch evd := respond.(type) {
	case error:
		response.WriteHeader(400)

		log.WithFields(logrus.Fields{
			"method":      request.Method,
			"request_uri": request.RequestURI,
			"form":        request.Form,
			"err":         evd,
		}).Warnln("respond failed")
	case string:
		response.Write([]byte(evd))
	default:
		response.WriteHeader(400)

		log.WithFields(logrus.Fields{
			"method":      request.Method,
			"request_uri": request.RequestURI,
			"form":        request.Form,
			"respond":     reflect.TypeOf(respond),
			"err":         evd,
		}).Warnln("respond unknown")
	}
}

func (w *httpHandler) getTotalRoomInfo(response http.ResponseWriter, request *http.Request) {
	ch := make(chan interface{})
	w.target.Tell(&hall_message.GetTotalRoom{
		Respond: func(payload string, err error) {
			if err != nil {
				ch <- err
			} else {
				ch <- payload
			}
			close(ch)
		},
	})
	respond := <-ch
	switch evd := respond.(type) {
	case error:
		response.WriteHeader(400)

		log.WithFields(logrus.Fields{
			"method":      request.Method,
			"request_uri": request.RequestURI,
			"form":        request.Form,
			"err":         evd,
		}).Warnln("respond failed")
	case string:
		response.Write([]byte(evd))
	default:
		response.WriteHeader(400)

		log.WithFields(logrus.Fields{
			"method":      request.Method,
			"request_uri": request.RequestURI,
			"form":        request.Form,
			"respond":     reflect.TypeOf(respond),
			"err":         evd,
		}).Warnln("respond unknown")
	}
}

func (w *httpHandler) playerChanged(response http.ResponseWriter, request *http.Request) {
	player := request.Form.Get("player_id")
	if player == "" {
		response.WriteHeader(400)
		return
	}
	id, err := strconv.ParseInt(player, 10, 64)
	if err != nil {
		response.WriteHeader(400)
		return
	}
	database.RefreshPlayer(database.Player(id))
	response.WriteHeader(200)
}

// 消息转发目标创建器
type TargetCreator func() *actor.PID

// 配置
type Option struct {
	// 创建器
	TargetCreator TargetCreator

	// 监听地址
	Address string
}

func Start(option Option) {
	l, err := net.Listen("tcp", option.Address)
	if err != nil {
		log.WithFields(logrus.Fields{
			"address": option.Address,
			"err":     err,
		}).Fatalln("listen failed")
	}

	go func() {
		err := http.Serve(l, &httpHandler{
			target: option.TargetCreator(),
		})
		if err != nil {
			log.WithFields(logrus.Fields{
				"address": option.Address,
				"err":     err,
			}).Fatalln("listen failed")
		}
	}()

	log.WithFields(logrus.Fields{
		"address": option.Address,
	}).Infoln("listen started")
}
