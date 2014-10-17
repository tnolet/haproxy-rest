package main
import (
	"github.com/Shopify/sarama"
	"strconv"
	"encoding/json"
	"time"
	"strings"
	"reflect"
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

	// The list of metrics we want to filter out of the total stats dump from haproxy
	wantedMetrics  := []string{ "Scur", "Qcur","Smax","Slim","Status","Weight","Qtime","Ctime","Rtime","Ttime","Req_rate","Req_rate_max","Req_tot","Rate","Rate_lim","Rate_max" }

	// get metrics every second, for ever.
	for  {

			stats, _ := GetStats("all")
		    localTime := int64(time.Now().Unix())


		// for each proxy in the stats dump, pick out the wanted metrics, parse them and send 'm to Kafka
			for _,proxy := range stats {

				// filter out the metrics for haproxy's own stats page
				if (proxy.Pxname != "stats") {

					// loop over all wanted metrics for the current proxy
					for _,metric := range wantedMetrics {

						fullMetricName := proxy.Pxname + "." + strings.ToLower(proxy.Svname) + "." + strings.ToLower(metric)
						field  := reflect.ValueOf(proxy).FieldByName(metric)
						metricValue := field.String()
						if (metricValue == "") { metricValue = "0"}
						//log.Info( localTime + " " + fullMetricName + " : " + metricValue)

						metricObj := Metric{fullMetricName, metricValue, localTime}
						jsonObj,_ := json.MarshalIndent(metricObj,""," ")

						err := producer.SendMessage(mode + "." + "all", sarama.StringEncoder("lbmetrics"), sarama.StringEncoder(jsonObj))
						if err != nil {

							log.Error("Error sending message to Kafka " + err.Error())

						} else {
							log.Debug("Successfully sent message to Kafka on topic: "  + mode + "." + "all")
						}


					}

				}


			}





		time.Sleep(3000 * time.Millisecond)
	}
}
