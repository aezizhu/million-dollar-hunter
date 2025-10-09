package kafka

import (
	"context"
	"log"
	"strings"

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
	ctx := context.Background()
	handler := &groupHandler{svc: c.svc, topic: c.cfg.TopicTxIngested}
	for {
		select {
		case <-c.stopCh:
			return
		default:
			if err := c.grp.Consume(ctx, []string{c.cfg.TopicTxIngested}, handler); err != nil {
				log.Printf("consume error: %v", err)
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
		if err := h.svc.HandleTransactionDataIngested(sess.Context(), msg.Value); err != nil {
			log.Printf("handle message err: %v", err)
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}
