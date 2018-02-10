package backend

import (
	"net"
	"net/http"
	"strconv"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka-cow/database"
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
		case "/playerChanged":
			w.playerChanged(response, request)
		case "/configurationChanged":
			w.configurationChanged(response, request)
		default:
			response.WriteHeader(405)
		}
	default:
		response.WriteHeader(405)
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

func (w *httpHandler) configurationChanged(response http.ResponseWriter, request *http.Request) {
	database.RefreshConfiguration()
	response.WriteHeader(200)
}

func StartHttp(option Option) {
	l, err := net.Listen("tcp", option.HttpAddress)
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
