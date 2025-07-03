package consumer

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-kratos/kratos/v2/log"
	kesselv2 "github.com/project-kessel/inventory-api/api/kessel/inventory/v1beta2"
	"github.com/project-kessel/inventory-consumer/consumer/auth"
	"github.com/project-kessel/inventory-consumer/consumer/retry"
	"github.com/project-kessel/inventory-consumer/consumer/transforms"
	kessel "github.com/project-kessel/inventory-consumer/internal/client"
	metricscollector "github.com/project-kessel/inventory-consumer/metrics"
	"go.opentelemetry.io/otel/attribute"
)

const (
	// commitModulo is used to define the batch size of offsets based on the current offset being processed
	commitModulo = 10

	OperationTypeCreated   = "created"
	OperationTypeUpdated   = "updated"
	OperationTypeDeleted   = "deleted"
	OperationTypeMigration = "migration"
)

var (
	requiredHeaders = []string{"operation"}
	ErrClosed       = errors.New("consumer closed")
	ErrMaxRetries   = errors.New("max retries reached")
)

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
	Consumer         Consumer
	Client           kessel.ClientProvider
	OffsetStorage    []kafka.TopicPartition
	Config           CompletedConfig
	MetricsCollector *metricscollector.MetricsCollector
	Logger           *log.Helper
	AuthOptions      *auth.Options
	RetryOptions     *retry.Options
}

// New instantiates a new InventoryConsumer
func New(config CompletedConfig, client kessel.ClientProvider, logger *log.Helper) (InventoryConsumer, error) {
	logger.Info("Setting up kafka consumer")
	logger.Debugf("completed kafka config: %+v", config.KafkaConfig)
	consumer, err := kafka.NewConsumer(config.KafkaConfig)
	if err != nil {
		logger.Errorf("error creating kafka consumer: %v", err)
		return InventoryConsumer{}, err
	}

	var mc metricscollector.MetricsCollector
	err = mc.New(config.Topics)
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
	Payload       KeyPayloadData         `json:"payload"`
}

type KeyPayloadData struct {
	ID string `json:"id"`
}

// MessagePayload stores the event message value captured from the topic as emitted by Debezium
type MessagePayload struct {
	MessageSchema  map[string]interface{} `json:"schema"`
	RequestPayload interface{}            `json:"payload"`
}

// Run starts the Consume loop with retry configurations and backoff
func (i *InventoryConsumer) Run(options *Options, config CompletedConfig, client kessel.ClientProvider, logger *log.Helper) error {
	retries := 0
	for options.RetryOptions.ConsumerMaxRetries == -1 || retries < options.RetryOptions.ConsumerMaxRetries {
		// If the consumer cannot process a message, the consumer loop is restarted
		// This is to ensure we re-read the message and prevent it being dropped and moving to next message.
		// To re-read the current message, we have to recreate the consumer connection so that the earliest offset is used
		kic, err := New(config, client, logger)
		if err != nil {
			return err
		}
		err = kic.Consume()
		if errors.Is(err, ErrClosed) {
			kic.Logger.Errorf("consumer unable to process current message -- restarting consumer")
			retries++
			if options.RetryOptions.ConsumerMaxRetries == -1 || retries < options.RetryOptions.ConsumerMaxRetries {
				backoff := min(time.Duration(kic.RetryOptions.BackoffFactor*retries*300)*time.Millisecond, time.Duration(options.RetryOptions.MaxBackoffSeconds)*time.Second)
				kic.Logger.Errorf("retrying in %v", backoff)
				time.Sleep(backoff)
			}
			continue
		} else {
			kic.Logger.Errorf("consumer unable to process messages: %v", err)
			return err
		}
	}
	return nil
}

// Consume begins the consumption loop for the Consumer
func (i *InventoryConsumer) Consume() error {
	err := i.Consumer.SubscribeTopics(i.Config.Topics, i.RebalanceCallback)
	if err != nil {
		metricscollector.Incr(i.MetricsCollector.ConsumerErrors, "SubscribeTopics", err)
		i.Logger.Errorf("failed to subscribe to topic: %v", err)
		return err
	}
	i.Logger.Infof("subscribed to topics: %s", strings.Join(i.Config.Topics, ", "))

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
				headers, err := ParseHeaders(e, requiredHeaders)
				if err != nil {
					metricscollector.Incr(i.MetricsCollector.MsgProcessFailures, "ParseHeaders", fmt.Errorf("missing headers"))
					i.Logger.Errorf("failed to parse message headers: %v", err)
					run = false
					continue
				}
				operation := headers["operation"]

				err = i.ProcessMessage(operation, e)
				if err != nil {
					i.Logger.Errorf(
						"error processing message: topic=%s partition=%d offset=%s",
						*e.TopicPartition.Topic, e.TopicPartition.Partition, e.TopicPartition.Offset)
					run = false
					continue
				}

				// store the current offset to be later batch committed
				i.OffsetStorage = append(i.OffsetStorage, e.TopicPartition)
				if CheckIfCommit(e.TopicPartition) {
					err := i.CommitStoredOffsets()
					if err != nil {
						metricscollector.Incr(i.MetricsCollector.ConsumerErrors, "CommitStoredOffsets", err)
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

// ProcessMessage processes an event message and replicates the change to Kessel Inventory
func (i *InventoryConsumer) ProcessMessage(operation string, msg *kafka.Message) error {
	switch operation {
	// TODO: We need to support migrations for many resource types, this is a temporary solution to support host migrations
	case OperationTypeMigration:
		i.Logger.Infof("processing message: operation=%s", operation)
		i.Logger.Debugf("processed message=%s", msg.Value)

		req, err := transforms.TransformHostToReportResourceRequest(msg.Value)
		if err != nil {
			metricscollector.Incr(i.MetricsCollector.MsgProcessFailures, "TransformHostToReportResourceRequest", err)
			i.Logger.Errorf("failed to parse message for host: %v", err)
			return err
		}

		if i.Client.IsEnabled() {
			resp, err := i.Retry(func() (interface{}, error) {
				return i.Client.CreateOrUpdateResource(req)
			})
			if err != nil {
				metricscollector.Incr(i.MetricsCollector.MsgProcessFailures, "CreateResource", err)
				i.Logger.Errorf("failed to create resource: %v", err)
				return err
			}
			i.Logger.Infof("response: %+v", resp)
		}
		return nil

	case OperationTypeCreated, OperationTypeUpdated:
		i.Logger.Infof("processing message: operation=%s", operation)
		i.Logger.Debugf("processed message=%s", msg.Value)

		var req kesselv2.ReportResourceRequest
		err := ParseCreateOrUpdateMessage(msg.Value, &req)
		if err != nil {
			metricscollector.Incr(i.MetricsCollector.MsgProcessFailures, "ParseCreateOrUpdateMessage", err)
			i.Logger.Errorf("failed to parse message for tuple: %v", err)
			return err
		}

		if i.Client.IsEnabled() {
			resp, err := i.Retry(func() (interface{}, error) {
				return i.Client.CreateOrUpdateResource(&req)
			})
			if err != nil {
				metricscollector.Incr(i.MetricsCollector.MsgProcessFailures, "CreateResource", err)
				i.Logger.Errorf("failed to create resource: %v", err)
				return err
			}
			i.Logger.Debugf("response: %v", resp)
		}
		return nil

	case OperationTypeDeleted:
		i.Logger.Infof("processing message: operation=%s", operation)
		i.Logger.Debugf("processed message=%s", msg.Value)

		var req kesselv2.DeleteResourceRequest
		err := ParseDeleteMessage(msg.Value, &req)
		if err != nil {
			metricscollector.Incr(i.MetricsCollector.MsgProcessFailures, "ParseDeleteMessage", err)
			i.Logger.Errorf("failed to parse message for filter: %v", err)
			return err
		}

		if i.Client.IsEnabled() {
			resp, err := i.Retry(func() (interface{}, error) {
				return i.Client.DeleteResource(&req)
			})
			if err != nil {
				metricscollector.Incr(i.MetricsCollector.MsgProcessFailures, "CreateResource", err)
				i.Logger.Errorf("failed to create resource: %v", err)
				return err
			}
			i.Logger.Debugf("response: %v", resp)
		}
		return nil

	default:
		metricscollector.Incr(i.MetricsCollector.MsgProcessFailures, "unknown-operation-type", nil)
		i.Logger.Errorf("unknown operation type, message cannot be processed and will be dropped: offset=%s operation=%s msg=%s",
			msg.TopicPartition.Offset.String(), operation, msg.Value)
	}
	return nil
}

// CheckIfCommit returns true whenever the condition to commit a batch of offsets is met
func CheckIfCommit(partition kafka.TopicPartition) bool {
	return partition.Offset%commitModulo == 0
}

// FormatOffsets converts a slice of partitions with offset data into a more readable shorthand-coded string to capture what partitions and offsets were comitted
func FormatOffsets(offsets []kafka.TopicPartition) string {
	var committedOffsets []string
	for _, partition := range offsets {
		committedOffsets = append(committedOffsets, fmt.Sprintf("[%d:%s]", partition.Partition, partition.Offset.String()))
	}
	return strings.Join(committedOffsets, ",")
}

// CommitStoredOffsets commits offsets for all processed messages since last offset commit
func (i *InventoryConsumer) CommitStoredOffsets() error {
	committed, err := i.Consumer.CommitOffsets(i.OffsetStorage)
	if err != nil {
		return err
	}

	i.Logger.Infof("offsets committed ([partition:offset]): %s", FormatOffsets(committed))
	i.OffsetStorage = nil
	return nil
}

// Shutdown ensures the consumer is properly shutdown, whether by server or due to rebalance
func (i *InventoryConsumer) Shutdown() error {
	if !i.Consumer.IsClosed() {
		i.Logger.Info("shutting down consumer...")
		if len(i.OffsetStorage) > 0 {
			err := i.CommitStoredOffsets()
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
func (i *InventoryConsumer) Retry(operation func() (interface{}, error)) (interface{}, error) {
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
		return resp, nil
	}
	i.Logger.Errorf("Error processing request (max attempts reached: %v): %v", attempts, err)
	return nil, ErrMaxRetries
}

// RebalanceCallback logs when rebalance events occur and ensures any stored offsets are committed before losing the partition assignment.
// It is registered to the kafka 'SubscribeTopics' call and is invoked automatically whenever rebalances occurs.
// Note, the RebalanceCb function must satisfy the function type func(*Consumer, Event).
// This function does so, but the consumer embedded in the InventoryConsumer is used versus the passed one which is the same consumer in either case.
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
		err := i.CommitStoredOffsets()
		if err != nil {
			i.Logger.Errorf("failed to commit offsets: %v", err)
			return err
		}

	default:
		i.Logger.Error("Unexpected event type: %v", event)
	}
	return nil
}
