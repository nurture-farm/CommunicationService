/*
 *  Copyright 2023 NURTURE AGTECH PVT LTD
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package kafka

import (
	"github.com/golang/protobuf/proto"
	"github.com/nurture-farm/CommunicationService/log"
	Common "github.com/nurture-farm/Contracts/Common/Gen/GoCommon"
	CommunicationEngine "github.com/nurture-farm/Contracts/CommunicationEngine/Gen/GoCommunicationEngine"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/confluentinc/confluent-kafka-go.v100/kafka"
	"time"
)

const (
	METRICS_NAME                      = "NF_CS_SEND_MESSAGE"
	HELP_MESSAGE                      = "Communication service metrics"
	LABELNAME_CODE                    = "code"
	LABEL_CODE_OK                     = "ok"
	LABEL_CODE_KO                     = "ko"
	MINUTES                           = 60 * 1000 * 1000 * 1000
	HIGH_PRIORITY_MESSAGE_EXPIRY_TIME = 5 * MINUTES
)

var producer *kafka.Producer

var logger = log.GetLogger()

var (
	sendMessageMetrics = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       METRICS_NAME,
		Help:       HELP_MESSAGE,
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{LABELNAME_CODE})
)

var lowPriorityTopicNameByChannel map[Common.CommunicationChannel]string

var highPriorityTopicNameByChannel map[Common.CommunicationChannel]string

func PushToRequestMetrics() func(*bool) {
	startTime := time.Now()
	return func(result *bool) {
		if *result == true {
			sendMessageMetrics.WithLabelValues(LABEL_CODE_OK).Observe(float64(time.Now().Sub(startTime).Milliseconds()))
		} else {
			sendMessageMetrics.WithLabelValues(LABEL_CODE_KO).Observe(float64(time.Now().Sub(startTime).Milliseconds()))
		}
	}
}

// InitProducer initializes the Kafka producer.
func InitProducer() {
	prometheus.MustRegister(sendMessageMetrics)
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":            viper.GetString("kafka.bootstrap.servers"),
		"acks":                         viper.GetInt("kafka.acks"),
		"compression.type":             viper.GetString("kafka.compression.type"),
		"compression.codec":            viper.GetString("kafka.compression.type"),
		"batch.num.messages":           cast.ToInt(viper.GetString("kafka.batch.num.messages")),
		"queue.buffering.max.ms":       cast.ToInt(viper.GetString("kafka.queue.buffering.max.ms")),
		"queue.buffering.max.messages": cast.ToInt(viper.GetString("kafka.queue.buffering.max.messages")),
	})

	if err != nil {
		logger.Panic("Failed to create kafka producer", zap.Error(err))
	}

	producer = p
	initHighPriorityTopicNameByChannelTypeMap()
	initLowPriorityTopicNameByChannelTypeMap()
}

// SendMessage sends the communication event to Kafka
func SendMessage(event *CommunicationEngine.CommunicationEvent) error {
	var ret error
	channelList := event.GetChannel()
	for index := range channelList {
		event.Channel = []Common.CommunicationChannel{channelList[index]}
		ret = sendMessageToKafka(event)
	}
	return ret
}

func sendMessageToKafka(event *CommunicationEngine.CommunicationEvent) error {
	doneChan := make(chan bool)
	var result bool
	defer PushToRequestMetrics()(&result)
	logger.Info("Sending communication", zap.Any("CommunicationEvent", event))
	go sendMessageCallback(doneChan, event)
	var topic = ""
	var isEventOfLowPriority = false
	if event.GetExpiry() != nil {
		isEventOfLowPriority = isLowPriority(event.GetExpiry())
	}
	if isEventOfLowPriority {
		topic = lowPriorityTopicNameByChannel[event.GetChannel()[0]]
	} else {
		topic = highPriorityTopicNameByChannel[event.GetChannel()[0]]
	}
	if len(topic) == 0 {
		topic = viper.GetString("communication.event.topics.default")
	}
	valueBytes, err := proto.Marshal(event)
	if err != nil {
		logger.Error("Failed to marshal event value", zap.Error(err))
		return err
	}

	var keyBytes []byte
	if event.ReceiverActor != nil {
		keyBytes, err = proto.Marshal(event.ReceiverActor)
		if err != nil {
			logger.Error("Failed to marshal event key", zap.Error(err))
			return err
		}
	} else {
		keyBytes, err = proto.Marshal(event.ReceiverActorDetails)
		if err != nil {
			logger.Error("Failed to marshal event key", zap.Error(err))
			return err
		}
	}

	message := kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          valueBytes,
		Key:            keyBytes,
	}
	producer.ProduceChannel() <- &message

	// wait for delivery report goroutine to finish
	result = <-doneChan
	return nil
}

func sendMessageCallback(doneChan chan bool, event *CommunicationEngine.CommunicationEvent) {
	defer close(doneChan)
	for e := range producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			m := ev
			if m.TopicPartition.Error != nil {
				logger.Error("Kafka producer failed to deliver message : ", zap.String("topic", *m.TopicPartition.Topic),
					zap.String("message_key", string(m.Key)), zap.Error(m.TopicPartition.Error), zap.Any("CommunicationEvent", event))
				doneChan <- false
			} else {
				logger.Info("Kafka producer successfully delivered message", zap.String("topic", *m.TopicPartition.Topic),
					zap.Int32("partition", m.TopicPartition.Partition), zap.Any("offset", m.TopicPartition.Offset), zap.Any("CommunicationEvent", event))
				doneChan <- true
			}
			return

		default:
			logger.Error("Ignored event: ", zap.Any("event", ev))
		}
	}
}

func initLowPriorityTopicNameByChannelTypeMap() {
	lowPriorityTopicNameByChannel = make(map[Common.CommunicationChannel]string)
	lowPriorityTopicNameByChannel[Common.CommunicationChannel_NO_CHANNEL] = viper.GetString("communication.event.topics.default")
	lowPriorityTopicNameByChannel[Common.CommunicationChannel_SMS] = viper.GetString("communication.event.topics.sms.low.priority")
	lowPriorityTopicNameByChannel[Common.CommunicationChannel_EMAIL] = viper.GetString("communication.event.topics.email.low.priority")
	lowPriorityTopicNameByChannel[Common.CommunicationChannel_APP_NOTIFICATION] = viper.GetString("communication.event.topics.pn.low.priority")
	lowPriorityTopicNameByChannel[Common.CommunicationChannel_WHATSAPP] = viper.GetString("communication.event.topics.whatsapp.low.priority")
}
func initHighPriorityTopicNameByChannelTypeMap() {
	highPriorityTopicNameByChannel = make(map[Common.CommunicationChannel]string)
	highPriorityTopicNameByChannel[Common.CommunicationChannel_NO_CHANNEL] = viper.GetString("communication.event.topics.default")
	highPriorityTopicNameByChannel[Common.CommunicationChannel_SMS] = viper.GetString("communication.event.topics.sms.high.priority")
	highPriorityTopicNameByChannel[Common.CommunicationChannel_EMAIL] = viper.GetString("communication.event.topics.email.high.priority")
	highPriorityTopicNameByChannel[Common.CommunicationChannel_APP_NOTIFICATION] = viper.GetString("communication.event.topics.pn.high.priority")
	highPriorityTopicNameByChannel[Common.CommunicationChannel_WHATSAPP] = viper.GetString("communication.event.topics.whatsapp.high.priority")
}

func isLowPriority(t *timestamppb.Timestamp) bool {
	return t.AsTime().Local().After(time.Now().Add(time.Duration(HIGH_PRIORITY_MESSAGE_EXPIRY_TIME)))
}
