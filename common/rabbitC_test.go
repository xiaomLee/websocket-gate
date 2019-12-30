package common

import (
	"fmt"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

func TestInitRabbit(t *testing.T) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s%s", "dev", "dev", "127.0.0.1", "5672", "/dev")
	if err := InitRabbit(url); err != nil {
		t.Error(err)
		return
	}
	rabbit := GetRabbitInstance()
	rabbit.ExchangeDeclare("dev_exchange")
	infoQueue, err := rabbit.QueueDeclare("", "dev_exchange", "info")
	if err != nil {
		t.Error(err)
		return
	}
	errQueue, err := rabbit.QueueDeclare("error", "dev_exchange", "error")
	if err != nil {
		t.Error(err)
		return
	}
	warningQueue, err := rabbit.QueueDeclare("", "dev_exchange", "warning")
	if err != nil {
		t.Error(err)
		return
	}

	rabbit.Consume("info_log_1", infoQueue, infoLog)
	rabbit.Consume("info_log_2", infoQueue, infoLog)
	rabbit.Consume("error_log", errQueue, errorLog)
	rabbit.Consume("warning_log", warningQueue, warningLog)

	ticker := time.NewTicker(1 * time.Second)
	timer := time.After(100 * time.Second)
	for {
		select {
		case <-ticker.C:
			str := time.Now().String()
			rabbit.PushTransientMessage("dev_exchange", "info", []byte(str))
			rabbit.PushTransientMessage("dev_exchange", "error", []byte(str))
		case <-timer:
			return
			//println("restart success")
		}
	}
}

func infoLog(d amqp.Delivery) {
	fmt.Printf("info consumer[%s] msg:%s \n", d.ConsumerTag, string(d.Body))
}

func errorLog(d amqp.Delivery) {
	fmt.Printf("error consumer[%s] msg:%s \n", d.ConsumerTag, string(d.Body))
}

func warningLog(d amqp.Delivery) {
	fmt.Printf("error consumer[%s] msg:%s \n", d.ConsumerTag, string(d.Body))
}
