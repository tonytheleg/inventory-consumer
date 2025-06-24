package consumer

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/tonytheleg/inventory-consumer/consumer/auth"
	"github.com/tonytheleg/inventory-consumer/consumer/retry"
	metricscollector "github.com/tonytheleg/inventory-consumer/metrics"
	"go.opentelemetry.io/otel/attribute"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/project-kessel/inventory-client-go/v1beta1"
)

const (
	// commitModulo is used to define the batch size of offsets based on the current offset being processed
	commitModulo         = 10
	OperationTypeCreated = "created"
	OperationTypeUpdated = "updated"
	OperationTypeDeleted = "deleted"
)

// defines all required headers for message processing
var requiredHeaders = []string{"operation"}

var ErrClosed = errors.New("consumer closed")
var ErrMaxRetries = errors.New("max retries reached")

type Consumer interface {
	CommitOffsets(offsets []kafka.TopicPartition) ([]kafka.TopicPartition, error)
	SubscribeTopics(topics []string, rebalanceCb kafka.RebalanceCb) (err error)
	Poll(timeoutMs int) (event kafka.Event)
	IsClosed() bool
	Close() error
	AssignmentLost() bool
}

// InventoryConsumer defines a Consumer with required clients and configs to call Relations API and update the Inventory DB with consistency tokens
type InventoryConsumer struct {
	Consumer      Consumer
	Client        *v1beta1.InventoryClient
	OffsetStorage []kafka.TopicPartition
	Config        CompletedConfig
	// AuthzConfig      authz.CompletedConfig	// Sets up Kessl AuthZ
	// Authorizer       api.Authorizer			// needs to be replaced with inventory client
	MetricsCollector *metricscollector.MetricsCollector
	Logger           *log.Helper
	AuthOptions      *auth.Options
	RetryOptions     *retry.Options
	// Notifier     pubsub.Notifier 			// R-A-W out of scope for external consumer for now
}

// New instantiates a new InventoryConsumer
func New(config CompletedConfig, client *v1beta1.InventoryClient, logger *log.Helper) (InventoryConsumer, error) {
	logger.Info("Setting up kafka consumer")
	logger.Debugf("completed kafka config: %+v", config.KafkaConfig)
	consumer, err := kafka.NewConsumer(config.KafkaConfig)
	if err != nil {
		logger.Errorf("error creating kafka consumer: %v", err)
		return InventoryConsumer{}, err
	}

	var mc metricscollector.MetricsCollector
	err = mc.New()
	if err != nil {
		logger.Errorf("error creating metrics collector: %v", err)
		return InventoryConsumer{}, err
	}

	authnOptions := &auth.Options{
		Enabled:          config.AuthConfig.Enabled,
		SecurityProtocol: config.AuthConfig.SecurityProtocol,
		SASLMechanism:    config.AuthConfig.SASLMechanism,
		SASLUsername:     config.AuthConfig.SASLUsername,
		SASLPassword:     config.AuthConfig.SASLPassword,
	}

	retryOptions := &retry.Options{
		ConsumerMaxRetries:  config.RetryConfig.ConsumerMaxRetries,
		OperationMaxRetries: config.RetryConfig.OperationMaxRetries,
		BackoffFactor:       config.RetryConfig.BackoffFactor,
		MaxBackoffSeconds:   config.RetryConfig.MaxBackoffSeconds,
	}

	return InventoryConsumer{
		Consumer:         consumer,
		Client:           client,
		OffsetStorage:    make([]kafka.TopicPartition, 0),
		Config:           config,
		MetricsCollector: &mc,
		Logger:           logger,
		AuthOptions:      authnOptions,
		RetryOptions:     retryOptions,
	}, nil
}

// KeyPayload stores the event message key captured from the topic as emitted by Debezium
type KeyPayload struct {
	MessageSchema map[string]interface{} `json:"schema"`
	InventoryID   string                 `json:"payload"`
}

// MessagePayload stores the event message value captured from the topic as emitted by Debezium
type MessagePayload struct {
	MessageSchema    map[string]interface{} `json:"schema"`
	RelationsRequest interface{}            `json:"payload"`
}

// Run starts the Consume loop with retry configurations and backoff
func (i *InventoryConsumer) Run(options *Options, config CompletedConfig, client *v1beta1.InventoryClient, logger *log.Helper) error {
	retries := 0
	for options.RetryOptions.ConsumerMaxRetries == -1 || retries < options.RetryOptions.ConsumerMaxRetries {
		// If the consumer cannot process a message, the consumer loop is restarted
		// This is to ensure we re-read the message and prevent it being dropped and moving to next message.
		// To re-read the current message, we have to recreate the consumer connection so that the earliest offset is used
		icrg, err := New(config, client, logger)
		if err != nil {
			return err
		}
		err = icrg.Consume()
		if errors.Is(err, ErrClosed) {
			icrg.Logger.Errorf("consumer unable to process current message -- restarting consumer")
			retries++
			if options.RetryOptions.ConsumerMaxRetries == -1 || retries < options.RetryOptions.ConsumerMaxRetries {
				backoff := min(time.Duration(icrg.RetryOptions.BackoffFactor*retries*300)*time.Millisecond, time.Duration(options.RetryOptions.MaxBackoffSeconds)*time.Second)
				icrg.Logger.Errorf("retrying in %v", backoff)
				time.Sleep(backoff)
			}
			continue
		} else {
			icrg.Logger.Errorf("consumer unable to process messages: %v", err)
			return err
		}
	}
	return nil
}

// Consume begins the consumption loop for the Consumer
func (i *InventoryConsumer) Consume() error {
	err := i.Consumer.SubscribeTopics([]string{i.Config.Topic}, i.RebalanceCallback)
	if err != nil {
		metricscollector.Incr(i.MetricsCollector.ConsumerErrors, "SubscribeTopics", err)
		i.Logger.Errorf("failed to subscribe to topic: %v", err)
		return err
	}

	// Set up a channel for handling exiting pods or ctrl+c
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Process messages
	run := true
	i.Logger.Info("Consumer ready: waiting for messages...")
	for run {
		select {
		case <-sigchan:
			run = false
		default:
			event := i.Consumer.Poll(100)
			if event == nil {
				continue
			}

			switch e := event.(type) {
			case *kafka.Message:
				headers, err := ParseHeaders(e)
				if err != nil {
					metricscollector.Incr(i.MetricsCollector.MsgProcessFailures, "ParseHeaders", fmt.Errorf("missing headers"))
					i.Logger.Errorf("failed to parse message headers: %v", err)
					run = false
					continue
				}
				operation := headers["operation"]

				//var resp interface{}

				_, err = i.ProcessMessage(headers, e)
				if err != nil {
					i.Logger.Errorf(
						"error processing message: topic=%s partition=%d offset=%s",
						*e.TopicPartition.Topic, e.TopicPartition.Partition, e.TopicPartition.Offset)
					run = false
					continue
				}

				if operation != OperationTypeDeleted {
					_, err := ParseMessageKey(e.Key)
					if err != nil {
						metricscollector.Incr(i.MetricsCollector.MsgProcessFailures, "ParseMessageKey", err)
						i.Logger.Errorf("failed to parse message key for for ID: %v", err)
					}
				}

				// store the current offset to be later batch committed
				i.OffsetStorage = append(i.OffsetStorage, e.TopicPartition)
				if checkIfCommit(e.TopicPartition) {
					err := i.commitStoredOffsets()
					if err != nil {
						metricscollector.Incr(i.MetricsCollector.ConsumerErrors, "commitStoredOffsets", err)
						i.Logger.Errorf("failed to commit offsets: %v", err)
						continue
					}
				}
				metricscollector.Incr(i.MetricsCollector.MsgsProcessed, operation, nil)
				i.Logger.Infof("consumed event from topic %s, partition %d at offset %s",
					*e.TopicPartition.Topic, e.TopicPartition.Partition, e.TopicPartition.Offset)
				i.Logger.Debugf("consumed event data: key = %-10s value = %s", string(e.Key), string(e.Value))

			case kafka.Error:
				metricscollector.Incr(i.MetricsCollector.KafkaErrorEvents, "kafka", nil,
					attribute.String("code", e.Code().String()),
					attribute.String("error", e.Error()))
				if e.IsFatal() {
					run = false
				} else {
					i.Logger.Errorf("recoverable consumer error: %v: %v -- will retry", e.Code(), e)
					continue
				}

			case *kafka.Stats:
				var stats metricscollector.StatsData
				err = json.Unmarshal([]byte(e.String()), &stats)
				if err != nil {
					metricscollector.Incr(i.MetricsCollector.MsgProcessFailures, "StatsCollection", err)
					i.Logger.Errorf("error unmarshalling stats: %v", err)
					continue
				}
				i.MetricsCollector.Collect(stats)
			default:
				i.Logger.Infof("event type ignored %v", e)
			}
		}
	}
	err = i.Shutdown()
	if !errors.Is(err, ErrClosed) {
		return fmt.Errorf("error in consumer shutdown: %v", err)
	}
	return err
}

func (i *InventoryConsumer) ProcessMessage(headers map[string]string, msg *kafka.Message) (string, error) {
	operation := headers["operation"]

	switch operation {
	case OperationTypeCreated:
		i.Logger.Infof("processing message: operation=%s", operation)
		i.Logger.Debugf("processed message tuple=%s", msg.Value)
		/* Convert to inventory calls
		   			tuple, err := ParseCreateOrUpdateMessage(msg.Value)
		    			if err != nil {
		   				metricscollector.Incr(i.MetricsCollector.MsgProcessFailures, "ParseCreateOrUpdateMessage", err)
		   				i.Logger.Errorf("failed to parse message for tuple: %v", err)
		   				return "", err
		   			}
		   			resp, err := i.Retry(func() (string, error) {
		   				return i.CreateTuple(context.Background(), tuple)
		   			})
		   			if err != nil {
		   				metricscollector.Incr(i.MetricsCollector.MsgProcessFailures, "CreateTuple", err)
		   				i.Logger.Errorf("failed to create tuple: %v", err)
		   				return "", err
		   			}
		   			return resp, nil
		   		}
		*/
	case OperationTypeUpdated:
		i.Logger.Infof("processing message: operation=%s", operation)
		i.Logger.Debugf("processed message tuple=%s", msg.Value)
		/* Convert to inventory calls
			tuple, err := ParseCreateOrUpdateMessage(msg.Value)
			if err != nil {
				metricscollector.Incr(i.MetricsCollector.MsgProcessFailures, "ParseCreateOrUpdateMessage", err)
				i.Logger.Errorf("failed to parse message for tuple: %v", err)
				return "", err
			}
			resp, err := i.Retry(func() (string, error) {
				return i.UpdateTuple(context.Background(), tuple)
			})
			if err != nil {
				metricscollector.Incr(i.MetricsCollector.MsgProcessFailures, "UpdateTuple", err)
				i.Logger.Errorf("failed to update tuple: %v", err)
				return "", err
			}
			return resp, nil
		}
		*/
	case OperationTypeDeleted:
		i.Logger.Infof("processing message: operation=%s", operation)
		i.Logger.Debugf("processed message tuple=%s", msg.Value)
		/* Convert to inventory calls
			filter, err := ParseDeleteMessage(msg.Value)
			if err != nil {
				metricscollector.Incr(i.MetricsCollector.MsgProcessFailures, "ParseDeleteMessage", err)
				i.Logger.Errorf("failed to parse message for filter: %v", err)
				return "", err
			}
			_, err = i.Retry(func() (string, error) {
				return i.DeleteTuple(context.Background(), filter)
			})
			if err != nil {
				metricscollector.Incr(i.MetricsCollector.MsgProcessFailures, "DeleteTuple", err)
				i.Logger.Errorf("failed to delete tuple: %v", err)
				return "", err
			}
			return "", nil
		}
		*/
	default:
		metricscollector.Incr(i.MetricsCollector.MsgProcessFailures, "unknown-operation-type", nil)
		i.Logger.Errorf("unknown operation type, message cannot be processed and will be dropped: offset=%s operation=%s msg=%s",
			msg.TopicPartition.Offset.String(), operation, msg.Value)
	}
	return "", nil
}

func ParseHeaders(msg *kafka.Message) (map[string]string, error) {
	headers := make(map[string]string)
	for _, v := range msg.Headers {
		// ignores any extra headers
		if slices.Contains(requiredHeaders, v.Key) {
			headers[v.Key] = string(v.Value)
		}
	}

	// ensures all required header keys are present after parsing, but only operation is required to have a value to process messages
	headerKeys := slices.Sorted(maps.Keys(headers))
	required := slices.Sorted(slices.Values(requiredHeaders))

	if !slices.Equal(headerKeys, required) || headers["operation"] == "" {
		return nil, fmt.Errorf("required headers are missing which would result in message processing failures: %+v", headers)
	}
	return headers, nil
}

//func ParseCreateOrUpdateMessage(msg []byte) (*v1beta1.Relationship, error) {
//	var msgPayload *MessagePayload
//	var tuple *v1beta1.Relationship
//
//	// msg value is expected to be a valid JSON body for a single relation
//	err := json.Unmarshal(msg, &msgPayload)
//	if err != nil {
//		return nil, fmt.Errorf("error unmarshaling msgPayload: %v", err)
//	}
//
//	payloadJson, err := json.Marshal(msgPayload.RelationsRequest)
//	if err != nil {
//		return nil, fmt.Errorf("error marshaling tuple payload: %v", err)
//	}
//
//	err = json.Unmarshal(payloadJson, &tuple)
//	if err != nil {
//		return nil, fmt.Errorf("error unmarshaling tuple payload: %v", err)
//	}
//	return tuple, nil
//}

//func ParseDeleteMessage(msg []byte) (*v1beta1.RelationTupleFilter, error) {
//	var msgPayload *MessagePayload
//	var filter *v1beta1.RelationTupleFilter
//
//	// msg value is expected to be a valid JSON body for a single relation
//	err := json.Unmarshal(msg, &msgPayload)
//	if err != nil {
//		return nil, fmt.Errorf("error unmarshaling msgPayload: %v", err)
//	}
//
//	payloadJson, err := json.Marshal(msgPayload.RelationsRequest)
//	if err != nil {
//		return nil, fmt.Errorf("error marshaling tuple payload: %v", err)
//	}
//
//	err = json.Unmarshal(payloadJson, &filter)
//	if err != nil {
//		return nil, fmt.Errorf("error unmarshaling tuple payload: %v", err)
//	}
//	return filter, nil
//}

func ParseMessageKey(msg []byte) (string, error) {
	var msgPayload *KeyPayload

	// msg key is expected to be the inventory_id of a resource
	err := json.Unmarshal(msg, &msgPayload)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling msgPayload: %v", err)
	}
	return msgPayload.InventoryID, nil
}

// checkIfCommit returns true whenever the condition to commit a batch of offsets is met
func checkIfCommit(partition kafka.TopicPartition) bool {
	return partition.Offset%commitModulo == 0
}

// formatOffsets converts a slice of partitions with offset data into a more readable shorthand-coded string to capture what partitions and offsets were comitted
func formatOffsets(offsets []kafka.TopicPartition) string {
	var committedOffsets []string
	for _, partition := range offsets {
		committedOffsets = append(committedOffsets, fmt.Sprintf("[%d:%s]", partition.Partition, partition.Offset.String()))
	}
	return strings.Join(committedOffsets, ",")
}

// commitStoredOffsets commits offsets for all processed messages since last offset commit
func (i *InventoryConsumer) commitStoredOffsets() error {
	committed, err := i.Consumer.CommitOffsets(i.OffsetStorage)
	if err != nil {
		return err
	}

	i.Logger.Infof("offsets committed ([partition:offset]): %s", formatOffsets(committed))
	i.OffsetStorage = nil
	return nil
}

// Shutdown ensures the consumer is properly shutdown, whether by server or due to rebalance
func (i *InventoryConsumer) Shutdown() error {
	if !i.Consumer.IsClosed() {
		i.Logger.Info("shutting down consumer...")
		if len(i.OffsetStorage) > 0 {
			err := i.commitStoredOffsets()
			if err != nil {
				i.Logger.Errorf("failed to commit offsets before shutting down: %v", err)
			}
		}
		err := i.Consumer.Close()
		if err != nil {
			i.Logger.Errorf("Error closing kafka consumer: %v", err)
			return err
		}
		return ErrClosed
	}
	return ErrClosed
}

// Retry executes the given function and will retry on failure with backoff until max retries is reached
func (i *InventoryConsumer) Retry(operation func() (string, error)) (string, error) {
	attempts := 0
	var resp interface{}
	var err error

	for i.RetryOptions.OperationMaxRetries == -1 || attempts < i.RetryOptions.OperationMaxRetries {
		resp, err = operation()
		if err != nil {
			metricscollector.Incr(i.MetricsCollector.MsgProcessFailures, "Retry", err)
			i.Logger.Errorf("request failed: %v", err)
			attempts++
			if i.RetryOptions.OperationMaxRetries == -1 || attempts < i.RetryOptions.OperationMaxRetries {
				backoff := min(time.Duration(i.RetryOptions.BackoffFactor*attempts*300)*time.Millisecond, time.Duration(i.RetryOptions.MaxBackoffSeconds)*time.Second)
				i.Logger.Errorf("retrying in %v", backoff)
				time.Sleep(backoff)
			}
			continue
		}
		return fmt.Sprintf("%s", resp), nil
	}
	i.Logger.Errorf("Error processing request (max attempts reached: %v): %v", attempts, err)
	return "", ErrMaxRetries
}

// RebalanceCallback logs when rebalance events occur and ensures any stored offsets are committed before losing the partition assignment. It is registered to the kafka 'SubscribeTopics' call and is invoked  automatically whenever rebalances occurs.
// Note, the RebalanceCb function must satisfy the function type func(*Consumer, Event). This function does so, but the consumer embedded in the InventoryConsumer is used versus the passed one which is the same consumer in either case.
func (i *InventoryConsumer) RebalanceCallback(consumer *kafka.Consumer, event kafka.Event) error {
	switch ev := event.(type) {
	case kafka.AssignedPartitions:
		i.Logger.Warnf("consumer rebalance event type: %d new partition(s) assigned: %v\n",
			len(ev.Partitions), ev.Partitions)

	case kafka.RevokedPartitions:
		i.Logger.Warnf("consumer rebalance event: %d partition(s) revoked: %v\n",
			len(ev.Partitions), ev.Partitions)

		if i.Consumer.AssignmentLost() {
			i.Logger.Warn("Assignment lost involuntarily, commit may fail")
		}
		err := i.commitStoredOffsets()
		if err != nil {
			i.Logger.Errorf("failed to commit offsets: %v", err)
			return err
		}

	default:
		i.Logger.Error("Unexpected event type: %v", event)
	}
	return nil
}
