package idgen_client

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type IdReq struct {
}

type IdRes struct {
	Id string `json:"id"`
}

func NewIdClient(connAddr, idtype string) *IdClient {
	return &IdClient{
		idtype:   idtype,
		connAddr: connAddr,
	}
}

type IdClient struct {
	connAddr string
	idtype   string
}

func (i *IdClient) BlockingGetId() (string, error) {
	conn, err := amqp091.Dial(i.connAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	// request exchange
	if err := ch.ExchangeDeclare("fim.idgen", "topic", true, false, false, false, nil); err != nil {
		panic(err)
	}

	// response queue
	queue, err := ch.QueueDeclare("fim.idgen_result_"+i.idtype, true, false, false, false, nil)
	if err != nil {
		panic(err)
	}
	if err := ch.QueueBind(queue.Name, "idtype."+i.idtype, "fim.idgen_result", false, nil); err != nil {
		panic(err)
	}

	data, err := json.Marshal(new(IdReq))
	if err != nil {
		return "", err
	}
	for retryCnt := 0; retryCnt < 3; retryCnt++ {
		// mandatory to make sure server is online
		if err := ch.PublishWithContext(context.Background(), "fim.idgen", "idtype."+i.idtype, true, false, amqp091.Publishing{
			ContentType: "application/json",
			Body:        data,
		}); err != nil {
			return "", err
		}

		// wait for processing
		time.Sleep(1 * time.Second)

		msg, ok, err := ch.Get(queue.Name, true)
		if err != nil {
			return "", err
		}
		if !ok {
			log.Println("didn't get id result, just retry!")
			continue
		}
		data := msg.Body
		res := new(IdRes)
		if err := json.Unmarshal(data, res); err != nil {
			return "", err
		}
		return res.Id, nil
	}
	return "", errors.New("reach maximum retry count")
}
