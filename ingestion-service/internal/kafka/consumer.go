package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/rs/zerolog"
)

type Consumer struct {
	groupID string
	topic   string
	grp     sarama.ConsumerGroup
	logger  zerolog.Logger
	handler WalletRequestHandler
	stopCh  chan struct{}
}

type WalletRequestHandler interface {
	HandleWalletTrackingRequest(ctx context.Context, event WalletTrackingRequestedEvent) error
}

type WalletTrackingRequestedEvent struct {
	EventID       string    `json:"event_id"`
	Timestamp     time.Time `json:"timestamp"`
	UserID        string    `json:"user_id"`
	WalletAddress string    `json:"wallet_address"`
	Chain         string    `json:"chain"`
	Nickname      string    `json:"nickname"`
}

func NewConsumer(ctx context.Context, brokers []string, groupID string, topic string, logger zerolog.Logger, handler WalletRequestHandler) (*Consumer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("at least one broker required")
	}
	if groupID == "" {
		return nil, fmt.Errorf("groupID required")
	}
	if topic == "" {
		return nil, fmt.Errorf("topic required")
	}
	if handler == nil {
		return nil, fmt.Errorf("handler required")
	}

	config := sarama.NewConfig()
	config.Version = sarama.V3_6_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	config.Consumer.Return.Errors = true

	grp, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("create consumer group: %w", err)
	}

	logger.Info().
		Strs("brokers", brokers).
		Str("group_id", groupID).
		Str("topic", topic).
		Msg("kafka consumer initialized")

	return &Consumer{
		groupID: groupID,
		topic:   topic,
		grp:     grp,
		logger:  logger,
		handler: handler,
		stopCh:  make(chan struct{}),
	}, nil
}

func (c *Consumer) Run(ctx context.Context) {
	handler := &consumerGroupHandler{
		logger:  c.logger,
		handler: c.handler,
		topic:   c.topic,
	}

	for {
		select {
		case <-c.stopCh:
			c.logger.Info().Msg("consumer stopped via stop channel")
			return
		case <-ctx.Done():
			c.logger.Info().Msg("consumer stopped via context cancellation")
			return
		default:
			if err := c.grp.Consume(ctx, []string{c.topic}, handler); err != nil {
				if errors.Is(err, context.Canceled) {
					c.logger.Info().Msg("consumer context canceled")
					return
				}
				c.logger.Error().Err(err).Msg("consumer error, retrying")
				time.Sleep(time.Second)
			}
		}
	}
}

func (c *Consumer) Stop() error {
	c.logger.Info().Msg("stopping kafka consumer")
	close(c.stopCh)
	if err := c.grp.Close(); err != nil {
		return fmt.Errorf("close consumer group: %w", err)
	}
	return nil
}

type consumerGroupHandler struct {
	logger  zerolog.Logger
	handler WalletRequestHandler
	topic   string
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	h.logger.Info().Msg("consumer group rebalanced (setup)")
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	h.logger.Info().Msg("consumer group rebalanced (cleanup)")
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		ctx := enrichContextWithMessageMetadata(sess.Context(), msg)

		h.logger.Info().
			Str("topic", msg.Topic).
			Int32("partition", msg.Partition).
			Int64("offset", msg.Offset).
			Msg("processing kafka message")

		err := h.processMessage(ctx, msg)
		if err != nil {
			errorContext := fmt.Sprintf("topic=%s partition=%d offset=%d timestamp=%v headers=%v",
				msg.Topic, msg.Partition, msg.Offset, msg.Timestamp, formatHeaders(msg.Headers))

			if isPermanentError(err) {
				h.logger.Error().
					Err(err).
					Str("context", errorContext).
					Msg("PERMANENT_ERROR: Skipping malformed message")
				sess.MarkMessage(msg, "")
			} else if isTransientError(err) {
				h.logger.Error().
					Err(err).
					Str("context", errorContext).
					Msg("TRANSIENT_ERROR: Retryable error, triggering rebalance")
				return fmt.Errorf("transient error, triggering rebalance: %w", err)
			} else {
				h.logger.Error().
					Err(err).
					Str("context", errorContext).
					Msg("UNKNOWN_ERROR: Unexpected error, triggering rebalance")
				return fmt.Errorf("unknown error, triggering rebalance: %w", err)
			}
		} else {
			sess.MarkMessage(msg, "")
			h.logger.Info().
				Str("topic", msg.Topic).
				Int32("partition", msg.Partition).
				Int64("offset", msg.Offset).
				Msg("MESSAGE_PROCESSED: Successfully processed message")
		}
	}
	return nil
}

func (h *consumerGroupHandler) processMessage(ctx context.Context, msg *sarama.ConsumerMessage) error {
	if len(msg.Value) == 0 {
		return fmt.Errorf("empty payload")
	}

	var event WalletTrackingRequestedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("unmarshal event: %w", err)
	}

	if event.WalletAddress == "" {
		return fmt.Errorf("wallet_address is required")
	}
	if event.Chain == "" {
		return fmt.Errorf("chain is required")
	}

	return h.handler.HandleWalletTrackingRequest(ctx, event)
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
		"is required",
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
