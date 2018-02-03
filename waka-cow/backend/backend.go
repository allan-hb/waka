package backend

import (
	"context"
	"errors"
	"net"
	"os"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/liuhan907/waka/waka-cow/database"
	"github.com/liuhan907/waka/waka-cow/modules/hall/hall_message"
	"github.com/liuhan907/waka/waka-cow/proto"
)

var (
	defaultAgentOption = &database.SupervisorData{
		BonusRate:     30,
		BaseScores:    []int32{2, 5, 10, 20, 50, 100, 200, 500},
		MaxRoomNumber: 4,
	}
	log = logrus.WithFields(logrus.Fields{
		"pid":    os.Getpid(),
		"module": "cow.backend",
	})
)

type backendServer struct {
	target *actor.PID
}

func (srv *backendServer) ConfigurationChanged(ctx context.Context, ev *empty.Empty) (*empty.Empty, error) {
	log.Debugln("configuration changed")
	return &empty.Empty{}, database.RefreshConfiguration()
}

func (srv *backendServer) BlacklistChanged(ctx context.Context, ev *empty.Empty) (*empty.Empty, error) {
	log.Debugln("blacklist changed")
	return &empty.Empty{}, database.RefreshSupervisorRoomBlacklist()
}

func (srv *backendServer) SupervisorChanged(ctx context.Context, ev *waka.SupervisorCond) (*empty.Empty, error) {
	log.WithFields(logrus.Fields{
		"supervisor": ev.Supervisor,
	}).Debugln("supervisor changed")

	database.RefreshSupervisor(database.Supervisor(ev.Supervisor))
	return &empty.Empty{}, nil
}

func (srv *backendServer) PlayerChanged(ctx context.Context, ev *waka.PlayerCond) (*empty.Empty, error) {
	log.WithFields(logrus.Fields{
		"player": ev.Player,
	}).Debugln("player changed")

	database.RefreshPlayer(database.Player(ev.Player))
	return &empty.Empty{}, nil
}

func (srv *backendServer) GetSupervisorRoom(ctx context.Context, ev *waka.GetSupervisorRoomRequest) (*waka.GetRoomResponse, error) {
	ch := make(chan interface{})
	defer close(ch)

	srv.target.Tell(&hall_message.GetSupervisorRoom{
		Player: database.Player(ev.GetPlayer()),
		Respond: func(response *waka.GetRoomResponse, e error) {
			if e != nil {
				ch <- e
			} else {
				ch <- response
			}
		},
	})
	response := <-ch
	switch evd := response.(type) {
	case *waka.GetRoomResponse:
		return evd, nil
	case error:
		return nil, evd
	default:
		return nil, errors.New("unknown error")
	}
}

func (srv *backendServer) GetPlayerRoom(ctx context.Context, ev *empty.Empty) (*waka.GetRoomResponse, error) {
	ch := make(chan interface{})
	defer close(ch)

	srv.target.Tell(&hall_message.GetPlayerRoom{
		Respond: func(response *waka.GetRoomResponse, e error) {
			if e != nil {
				ch <- e
			} else {
				ch <- response
			}
		},
	})
	response := <-ch
	switch evd := response.(type) {
	case *waka.GetRoomResponse:
		return evd, nil
	case error:
		return nil, evd
	default:
		return nil, errors.New("unknown error")
	}
}

func (srv *backendServer) GetOnlinePlayer(ctx context.Context, ev *empty.Empty) (*waka.GetOnlinePlayerResponse, error) {
	ch := make(chan interface{})
	defer close(ch)

	srv.target.Tell(&hall_message.GetOnlinePlayer{
		Respond: func(response *waka.GetOnlinePlayerResponse, e error) {
			if e != nil {
				ch <- e
			} else {
				ch <- response
			}
		},
	})
	response := <-ch
	switch evd := response.(type) {
	case *waka.GetOnlinePlayerResponse:
		return evd, nil
	case error:
		return nil, evd
	default:
		return nil, errors.New("unknown error")
	}
}

func (srv *backendServer) KickPlayer(ctx context.Context, ev *waka.KickPlayerRequest) (*empty.Empty, error) {
	log.Debugln("kick player")
	srv.target.Tell(&hall_message.KickPlayer{Player: database.Player(ev.GetPlayer())})
	return &empty.Empty{}, nil
}

func (srv *backendServer) KickRoom(ctx context.Context, ev *waka.KickRoomRequest) (*empty.Empty, error) {
	log.Debugln("kick room")
	srv.target.Tell(&hall_message.KickRoom{Room: ev.GetRoomId()})
	return &empty.Empty{}, nil
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
	srv := grpc.NewServer()
	self := &backendServer{
		target: option.TargetCreator(),
	}
	waka.RegisterBackendServer(srv, self)
	go func() {
		err := srv.Serve(l)
		if err != nil {
			log.WithFields(logrus.Fields{
				"err": err,
			}).Fatalln("serve failed")
		}
	}()

	log.WithFields(logrus.Fields{
		"address": option.Address,
	}).Infoln("listen started")
}
