package kafka

import (
	"context"
	"errors"
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
		err := h.svc.HandleTransactionDataIngested(sess.Context(), msg.Value)
		if err != nil {
			log.Printf("handle message err: %v, topic=%s partition=%d offset=%d",
				err, msg.Topic, msg.Partition, msg.Offset)

			if isPermanentError(err) {
				log.Printf("permanent error, skipping message: topic=%s partition=%d offset=%d",
					msg.Topic, msg.Partition, msg.Offset)
				sess.MarkMessage(msg, "")
			} else {
				return err
			}
		} else {
			sess.MarkMessage(msg, "")
		}
	}
	return nil
}

func isPermanentError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "unmarshal") ||
		strings.Contains(errMsg, "invalid") ||
		strings.Contains(errMsg, "malformed")
}
