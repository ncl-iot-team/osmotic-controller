package srv

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

//RunCommander runs the main commander program
func RunCommander(fileName string) {

	file, err := os.Open(fileName)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer file.Close()

	reader := bufio.NewReader(file)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {

		line := scanner.Text()
		if strings.HasPrefix(line, "wait") {

			timeSecStr := strings.TrimLeft(line, "wait ")

			timeSec, err := strconv.ParseInt(timeSecStr, 10, 64)

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			time.Sleep(time.Second * time.Duration(timeSec))

		} else if strings.HasPrefix(strings.ToLower(line), "set e_fwd_rt") || strings.HasPrefix(strings.ToLower(line), "set s_rt") {

			var amqpConnDetails AMQPConnDetailsType

			amqpConnDetails.host = viper.GetString("remote-device1.host")
			amqpConnDetails.port = viper.GetString("remote-device1.port")
			amqpConnDetails.user = viper.GetString("remote-device1.user")
			amqpConnDetails.password = viper.GetString("remote-device1.password")
			amqpConnDetails.queue = viper.GetString("remote-device1.queue")

			fmt.Printf("%d | Sent Commmand : '%s'\n", time.Now().UnixNano(), line)

			SendCMD(amqpConnDetails, strings.ToLower(line))

		}

		//	deliveries := make(chan string, 4096)
		//	go NewProducer(amqpConnDetails, &deliveries)

	}

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

//SendCMD sends the command to command channel
func SendCMD(amqpConnDetails AMQPConnDetailsType, command string) {

	connStr := fmt.Sprintf("amqp://%s:%s@%s:%s", amqpConnDetails.user, amqpConnDetails.password, amqpConnDetails.host, amqpConnDetails.port)

	log.Printf("Connection String: %s", connStr)

	conn, err := amqp.Dial(connStr)

	failOnError(err, "Failed to connect to RabbitMQ")

	defer conn.Close()

	ch, err := conn.Channel()

	failOnError(err, "Failed to open a channel")

	defer ch.Close()

	q, err := ch.QueueDeclare(
		amqpConnDetails.queue, // name
		true,                  // durable
		false,                 // delete when unused
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	body := command
	err = ch.Publish("cmdr-exchange", // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})

	failOnError(err, "Failed to publish a message")

}

//SendSingleCMD sends a single command to command channel
func SendSingleCMD(command string) {

	var amqpConnDetails AMQPConnDetailsType

	amqpConnDetails.host = viper.GetString("remote-device1.host")
	amqpConnDetails.port = viper.GetString("remote-device1.port")
	amqpConnDetails.user = viper.GetString("remote-device1.user")
	amqpConnDetails.password = viper.GetString("remote-device1.password")
	amqpConnDetails.queue = viper.GetString("remote-device1.queue")

	SendCMD(amqpConnDetails, command)

}
