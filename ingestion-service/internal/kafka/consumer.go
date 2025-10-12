package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/rs/zerolog"
)

var (
	ethereumAddressRegex = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
	solanaAddressRegex   = regexp.MustCompile(`^[1-9A-HJ-NP-Za-km-z]{43,44}$`)
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

type TransientError struct {
	Underlying error
}

func (e *TransientError) Error() string {
	return fmt.Sprintf("transient error: %v", e.Underlying)
}

func (e *TransientError) Unwrap() error {
	return e.Underlying
}

type Consumer struct {
	groupID string
	topic   string
	grp     sarama.ConsumerGroup
	logger  zerolog.Logger
	handler WalletRequestHandler
	metrics ConsumerMetrics
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

type ConsumerMetrics interface {
	ObserveConsume(topic string, err error, start time.Time)
	ObserveRebalance()
}

type WalletRequestHandler interface {
	HandleWalletTrackingRequest(ctx context.Context, event WalletTrackingRequestedEvent) error
}

type WalletTrackingRequestedEvent struct {
	SchemaVersion string    `json:"schema_version"`
	EventID       string    `json:"event_id"`
	Timestamp     time.Time `json:"timestamp"`
	UserID        string    `json:"user_id"`
	WalletAddress string    `json:"wallet_address"`
	Chain         string    `json:"chain"`
	Nickname      string    `json:"nickname"`
}

var validChains = map[string]bool{
	"ethereum": true,
	"bsc":      true,
	"polygon":  true,
	"arbitrum": true,
	"optimism": true,
	"solana":   true,
}

func NewConsumer(ctx context.Context, brokers []string, groupID string, topic string, logger zerolog.Logger, handler WalletRequestHandler, metrics ConsumerMetrics) (*Consumer, error) {
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
		metrics: metrics,
		stopCh:  make(chan struct{}),
	}, nil
}

func (c *Consumer) Run(ctx context.Context) {
	c.wg.Add(1)
	defer c.wg.Done()

	handler := &consumerGroupHandler{
		logger:  c.logger,
		handler: c.handler,
		metrics: c.metrics,
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
	c.wg.Wait()
	if err := c.grp.Close(); err != nil {
		return fmt.Errorf("close consumer group: %w", err)
	}
	return nil
}

type consumerGroupHandler struct {
	logger  zerolog.Logger
	handler WalletRequestHandler
	metrics ConsumerMetrics
	topic   string
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	h.logger.Info().Msg("consumer group rebalanced (setup)")
	if h.metrics != nil {
		h.metrics.ObserveRebalance()
	}
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	h.logger.Info().Msg("consumer group rebalanced (cleanup)")
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		start := time.Now()
		ctx := enrichContextWithMessageMetadata(sess.Context(), msg)

		h.logger.Info().
			Str("topic", msg.Topic).
			Int32("partition", msg.Partition).
			Int64("offset", msg.Offset).
			Msg("processing kafka message")

		err := h.processMessage(ctx, msg)
		
		if h.metrics != nil {
			h.metrics.ObserveConsume(msg.Topic, err, start)
		}
		
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
		return &ValidationError{Field: "payload", Message: "empty payload"}
	}

	var event WalletTrackingRequestedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("unmarshal event: %w", err)
	}

	if err := validateEvent(&event); err != nil {
		payloadPreview := string(msg.Value)
		if len(payloadPreview) > 100 {
			payloadPreview = payloadPreview[:100] + "..."
		}
		h.logger.Error().
			Err(err).
			Str("payload_preview", payloadPreview).
			Msg("event validation failed")
		return err
	}

	return h.handler.HandleWalletTrackingRequest(ctx, event)
}

func validateEvent(event *WalletTrackingRequestedEvent) error {
	if event.SchemaVersion == "" {
		return &ValidationError{Field: "schema_version", Message: "required field is empty"}
	}
	if event.SchemaVersion != "1.0.0" {
		return &ValidationError{
			Field:   "schema_version",
			Message: fmt.Sprintf("unsupported schema version '%s', expected '1.0.0'", event.SchemaVersion),
		}
	}

	if event.WalletAddress == "" {
		return &ValidationError{Field: "wallet_address", Message: "required field is empty"}
	}

	if event.Chain == "" {
		return &ValidationError{Field: "chain", Message: "required field is empty"}
	}

	if !validChains[event.Chain] {
		return &ValidationError{
			Field:   "chain",
			Message: fmt.Sprintf("invalid chain '%s', must be one of: ethereum, bsc, polygon, arbitrum, optimism, solana", event.Chain),
		}
	}

	if err := validateWalletAddress(event.WalletAddress, event.Chain); err != nil {
		return err
	}

	return nil
}

func validateWalletAddress(address, chain string) error {
	switch chain {
	case "ethereum", "bsc", "polygon", "arbitrum", "optimism":
		if !ethereumAddressRegex.MatchString(address) {
			return &ValidationError{
				Field:   "wallet_address",
				Message: fmt.Sprintf("invalid EVM address format: must be 42 characters starting with 0x (got: %s)", address),
			}
		}
	case "solana":
		if !solanaAddressRegex.MatchString(address) {
			return &ValidationError{
				Field:   "wallet_address",
				Message: fmt.Sprintf("invalid Solana address format: must be 43-44 base58 characters (got: %s)", address),
			}
		}
	}
	return nil
}

func isPermanentError(err error) bool {
	if err == nil {
		return false
	}

	var valErr *ValidationError
	if errors.As(err, &valErr) {
		return true
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return true
	}

	errMsg := strings.ToLower(err.Error())
	permanentKeywords := []string{
		"unmarshal",
		"malformed",
		"invalid format",
		"invalid json",
		"bad request",
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

	var transErr *TransientError
	if errors.As(err, &transErr) {
		return true
	}

	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "connection") ||
		strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "temporary") ||
		strings.Contains(errMsg, "unavailable") ||
		strings.Contains(errMsg, "deadline exceeded") ||
		strings.Contains(errMsg, "context deadline") ||
		strings.Contains(errMsg, "queue full")
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
