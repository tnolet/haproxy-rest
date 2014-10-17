package main
import (
	"github.com/Shopify/sarama"
	"strconv"
	"encoding/json"
	"time"
)


func setUpProducer(host string, port int, mode string) {

	connection := host + ":" + strconv.Itoa(port)

	log.Info("Connecting to Kafka on " + connection + "...")

	clientConfig := sarama.NewClientConfig()
	clientConfig.WaitForElection = (10 * time.Second)


	client, err := sarama.NewClient("client_id", []string{connection}, clientConfig)
	if err != nil {
		panic(err)
	} else {
		log.Info("Connection to Kafka successful")
	}

	/**
	*	 Create a producer with some specific setting
	*/
	producerConfig := sarama.NewProducerConfig()

	// if delivering messages async,  buffer them for at most MaxBufferTime
	producerConfig.MaxBufferTime = (2 * time.Second)

	// max bytes in buffer
	producerConfig.MaxBufferedBytes = 51200

	// Use zip compression
	producerConfig.Compression = 0

	// We are just streaming metrics, so don't not wait for any Kafka Acks.
	producerConfig.RequiredAcks = -1

	producer, err := sarama.NewProducer(client, producerConfig)
	if err != nil {
		panic(err)
	}

	go pushMetrics(producer,mode)

}

// pushMetrics pushes the load balancer statistic to a Kafka Topic
func pushMetrics(producer *sarama.Producer, mode string) {

	// loop over this collection of metric types
	metricTypes  := []string{ "all","frontend","backend","server" }


	for i := 0; i < 1; {

		for _,metricType := range metricTypes {

			stats, _ := GetStats(metricType)
			statsJson, _ := json.Marshal(stats)

			// prepend the metrics with the mode
			err := producer.SendMessage(mode + "." + metricType, nil, sarama.StringEncoder(statsJson))
			if err != nil {

				log.Error("Error sending message to Kafka " + err.Error())

			} else {
				log.Debug("Successfully sent message to Kafka on topic: " + metricType)
			}

		}
		time.Sleep(1000 * time.Millisecond)
	}
}
