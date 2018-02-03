package logger

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (my *actorT) ReceiveClock(context actor.Context) bool {
	switch context.Message().(type) {
	case *clock1:
		my.clock1()
	default:
		return false
	}
	return true
}

// ---------------------------------------------------------------------------------------------------------------------

type clock1 struct{}

func (my *actorT) clock1() {
	defer func() {
		time.AfterFunc(time.Second, func() { my.pid.Tell(&clock1{}) })
	}()

	if my.w.Len() == 0 {
		return
	}

	fileName := fmt.Sprintf("%s_%s", my.option.Prefix, my.name)

	fd, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0)
	if err != nil {
		log.Printf("open log file \"%s\" failed: %v\n", fileName, err)
		return
	}
	defer fd.Close()

	_, err = fd.Write(my.w.Bytes())
	if err != nil {
		log.Printf("write log file failed: %v\n", err)
		return
	}

	my.w.Reset()
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) startClock() {
	my.clock1()
}
