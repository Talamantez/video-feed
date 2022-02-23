package kafka

import (
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gyozatech/noodlog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var nlog *noodlog.Logger

var producer *kafka.Producer

func InitProducer() {

	broker := "localhost:9092"

	var err error
	producer, err = kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": broker})

	if err != nil {
		log.Print("Failed to create producer")
		panic(err)
	}

	// log.Printf("Created Producer %v", producer)

}
func init() {
	nlog = noodlog.NewLogger().SetConfigs(
		noodlog.Configs{
			LogLevel:             noodlog.LevelTrace,
			JSONPrettyPrint:      noodlog.Enable,
			TraceCaller:          noodlog.Enable,
			Colors:               noodlog.Enable,
			CustomColors:         &noodlog.CustomColors{Trace: noodlog.Cyan},
			ObscureSensitiveData: noodlog.Enable,
			SensitiveParams:      []string{"password"},
		},
	)
}

func Produce(topic string, value string) {
	start := time.Now()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	deliveryChan := make(chan kafka.Event)

	err := producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(value),
		Headers:        []kafka.Header{{Key: "InboundTopic", Value: []byte("InboundTopic feed header value")}},
	}, deliveryChan)

	e := <-deliveryChan

	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		log.Printf("Delivery failed: %v", m.TopicPartition.Error)
	} else {
		end := time.Now()
		duration := end.Sub(start)
		log := zerolog.New(os.Stdout).With().
			Timestamp().
			Str("app", "KafRedigo").Dur("Duration", duration).
			Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})

		log.Printf("Produced %s [%d] at offset %v",
			*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
		nlog.Trace("Successfully Produced")
	}
	if err != nil {
		log.Printf("Error in writing value : %v ", err)
	}
	close(deliveryChan)
}
