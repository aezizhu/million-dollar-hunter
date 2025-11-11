// Package kafka provides event consumption for portfolio updates.
// The consumer implementation handles TransactionDataIngested events from Kafka,
// processing them to maintain the read model with careful attention to message
// ordering, error handling, and consumer group coordination.
package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/aezizhu/million-dollar-hunter/services/portfolio-service/internal/config"
	"github.com/aezizhu/million-dollar-hunter/services/portfolio-service/internal/service"
)

type Consumer struct {
	cfg    config.Config
	svc    *service.PortfolioService
	grp    sarama.ConsumerGroup
	stopCh chan struct{}
}

func NewConsumer(cfg config.Config, svc *service.PortfolioService) (*Consumer, error) {
	c := sarama.NewConfig()
	c.Version = sarama.V3_6_0_0
	c.Consumer.Offsets.Initial = sarama.OffsetNewest

	grp, err := sarama.NewConsumerGroup(strings.Split(cfg.KafkaBrokers, ","), cfg.GroupID, c)
	if err != nil {
		return nil, err
	}
	return &Consumer{cfg: cfg, svc: svc, grp: grp, stopCh: make(chan struct{})}, nil
}

func (c *Consumer) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := &groupHandler{svc: c.svc, topic: c.cfg.TopicTxIngested}
	for {
		select {
		case <-c.stopCh:
			cancel()
			return
		default:
			if err := c.grp.Consume(ctx, []string{c.cfg.TopicTxIngested}, handler); err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				log.Printf("consume error: %v", err)
				time.Sleep(time.Second)
			}
		}
	}
}

func (c *Consumer) Stop() {
	close(c.stopCh)
	_ = c.grp.Close()
}

type groupHandler struct {
	svc   *service.PortfolioService
	topic string
}

func (h *groupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *groupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (h *groupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		ctx := enrichContextWithMessageMetadata(sess.Context(), msg)
		
		err := h.svc.HandleTransactionDataIngested(ctx, msg.Value)
		if err != nil {
			errorContext := fmt.Sprintf("topic=%s partition=%d offset=%d timestamp=%v headers=%v",
				msg.Topic, msg.Partition, msg.Offset, msg.Timestamp, formatHeaders(msg.Headers))
			
			if isPermanentError(err) {
				log.Printf("PERMANENT_ERROR: Skipping malformed message: %v | %s", err, errorContext)
				sess.MarkMessage(msg, "")
			} else if isTransientError(err) {
				log.Printf("TRANSIENT_ERROR: Retryable error encountered: %v | %s | Rebalancing consumer group", err, errorContext)
				return fmt.Errorf("transient error, triggering rebalance: %w", err)
			} else {
				log.Printf("UNKNOWN_ERROR: Unexpected error: %v | %s | Rebalancing consumer group", err, errorContext)
				return fmt.Errorf("unknown error, triggering rebalance: %w", err)
			}
		} else {
			sess.MarkMessage(msg, "")
			log.Printf("MESSAGE_PROCESSED: Successfully processed message | topic=%s partition=%d offset=%d",
				msg.Topic, msg.Partition, msg.Offset)
		}
	}
	return nil
}

func isPermanentError(err error) bool {
	if err == nil {
		return false
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return true
	}

	errMsg := strings.ToLower(err.Error())
	permanentKeywords := []string{
		"unmarshal",
		"malformed",
		"empty payload",
		"invalid format",
		"invalid json",
		"bad request",
		"validation failed",
	}

	for _, keyword := range permanentKeywords {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}

	return false
}

func isTransientError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "connection") ||
		strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "temporary") ||
		strings.Contains(errMsg, "unavailable") ||
		strings.Contains(errMsg, "deadline exceeded") ||
		strings.Contains(errMsg, "context deadline")
}

type contextKey string

const (
	kafkaTopicKey     contextKey = "kafka.topic"
	kafkaPartitionKey contextKey = "kafka.partition"
	kafkaOffsetKey    contextKey = "kafka.offset"
	kafkaTimestampKey contextKey = "kafka.timestamp"
)

func enrichContextWithMessageMetadata(ctx context.Context, msg *sarama.ConsumerMessage) context.Context {
	ctx = context.WithValue(ctx, kafkaTopicKey, msg.Topic)
	ctx = context.WithValue(ctx, kafkaPartitionKey, msg.Partition)
	ctx = context.WithValue(ctx, kafkaOffsetKey, msg.Offset)
	ctx = context.WithValue(ctx, kafkaTimestampKey, msg.Timestamp)
	return ctx
}

func formatHeaders(headers []*sarama.RecordHeader) string {
	if len(headers) == 0 {
		return "none"
	}
	var parts []string
	for _, h := range headers {
		parts = append(parts, fmt.Sprintf("%s=%s", string(h.Key), string(h.Value)))
	}
	return strings.Join(parts, ", ")
}
