package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/gofrs/uuid/v5"
	"github.com/rabbitmq/amqp091-go"
)

type IdReq struct {
}

type IdRes struct {
	Id string `json:"id"`
}

var connAdd string

func init() {
	flag.StringVar(&connAdd, "connAddr", "", "-connAddr=amqp://guest:guest@localhost")
}

func main() {
	flag.Parse()
	if connAdd == "" {
		panic(errors.New("connAddr is empty"))
	}

	conn, err := amqp091.Dial(connAdd)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	// request exchange-queue
	if err := ch.ExchangeDeclare("fim.idgen", "topic", true, false, false, false, nil); err != nil {
		panic(err)
	}
	queue, err := ch.QueueDeclare("fim.idgenserver", true, false, false, false, nil)
	if err != nil {
		panic(err)
	}
	if err := ch.QueueBind(queue.Name, "idtype.*", "fim.idgen", false, nil); err != nil {
		panic(err)
	}
	// response exchange with the same routing key as request's
	if err := ch.ExchangeDeclare("fim.idgen_result", "topic", true, false, false, false, nil); err != nil {
		panic(err)
	}

	msgCh, err := ch.Consume("fim.idgenserver", "", true, false, false, false, nil)
	if err != nil {
		panic(err)
	}

	// process request in parallel
	for i := 0; i < runtime.GOMAXPROCS(-1); i++ {
		go func() {
			for {
				msg := <-msgCh
				req := new(IdReq)
				if err := json.Unmarshal(msg.Body, req); err != nil {
					log.Printf("unmarshal msg body failed:%s", err)
					continue
				}
				log.Println("incoming id request for idtype:" + msg.RoutingKey)

				//FIXME need an algorithm like snowflake or others
				uuidInst, err := uuid.NewV4()
				if err != nil {
					log.Printf("generate uuid failed:%s", err)
					continue
				}

				res := IdRes{Id: uuidInst.String()}
				data, err := json.Marshal(res)
				if err != nil {
					log.Printf("marshal id response failed:%s", err)
					continue
				}

				// set message expiration of 10s
				if err := ch.PublishWithContext(context.Background(), "fim.idgen_result", msg.RoutingKey, false, false, amqp091.Publishing{
					Expiration:  "10000",
					ContentType: "application/json",
					Body:        data,
				}); err != nil {
					log.Printf("publish message failed:%s", err)
					continue
				}
			}
		}()
	}

	log.Println("service started!")

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT)
	_ = <-c
	log.Println("service exit!")
}
