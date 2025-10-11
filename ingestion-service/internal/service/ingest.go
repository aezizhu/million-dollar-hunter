package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/alchemy"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/breaker"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/config"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/kafka"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/metrics"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/models"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/moralis"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/ratelimit"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/repository"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/solana"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/transformer"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type Service struct {
	cfg            *config.Config
	logger         zerolog.Logger
	db             *repository.Postgres
	alc            *alchemy.Client
	mor            *moralis.Client
	sol            *solana.Client
	jobch          chan models.IngestionJob
	producer       *kafka.Producer
	kafkaMetrics   *metrics.KafkaMetrics
	ingestionMetrics *metrics.IngestionMetrics
}

func New(ctx context.Context, cfg *config.Config, logger zerolog.Logger, db *repository.Postgres, reg *prometheus.Registry) (*Service, error) {
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	alcBr := breaker.New("alchemy")
	morBr := breaker.New("moralis")
	solBr := breaker.New("solana")
	alcRl := ratelimit.New(rdb, "alchemy", "transfers", 20, 40)
	morRl := ratelimit.New(rdb, "moralis", "balances", 10, 20)
	solRl := ratelimit.New(rdb, "solana", "transactions", 15, 30)
	alc := alchemy.New(cfg.AlchemyBaseURL, cfg.AlchemyAPIKey, alcBr, alcRl)
	mor := moralis.New(cfg.MoralisBaseURL, cfg.MoralisAPIKey, morBr, morRl)
	sol := solana.New(cfg.MoralisBaseURL, cfg.MoralisAPIKey, solBr, solRl)

	var kafkaMetrics *metrics.KafkaMetrics
	var ingestionMetrics *metrics.IngestionMetrics
	if reg != nil {
		kafkaMetrics = metrics.NewKafkaMetrics(reg, "ingestion")
		ingestionMetrics = metrics.NewIngestionMetrics(reg, "ingestion")
	}

	var producer *kafka.Producer
	if cfg.KafkaEnabled {
		brokers := strings.Split(cfg.KafkaBrokers, ",")
		p, err := kafka.NewProducer(ctx, brokers, cfg.KafkaTopicTxIngested, logger, kafkaMetrics)
		if err != nil {
			logger.Error().Err(err).Msg("kafka producer init failed, continuing without kafka")
		} else {
			producer = p
		}
	}

	return &Service{
		cfg:              cfg,
		logger:           logger,
		db:               db,
		alc:              alc,
		mor:              mor,
		sol:              sol,
		jobch:            make(chan models.IngestionJob, 64),
		producer:         producer,
		kafkaMetrics:     kafkaMetrics,
		ingestionMetrics: ingestionMetrics,
	}, nil
}

func (s *Service) Enqueue(job models.IngestionJob) error {
	select {
	case s.jobch <- job:
		return nil
	default:
		s.logger.Warn().
			Str("wallet", job.Wallet).
			Str("chain", job.Chain).
			Int("queue_depth", len(s.jobch)).
			Msg("job queue full, rejecting job")
		return fmt.Errorf("job queue full (depth: %d/%d)", len(s.jobch), cap(s.jobch))
	}
}

func (s *Service) HandleWalletTrackingRequest(ctx context.Context, event kafka.WalletTrackingRequestedEvent) error {
	s.logger.Info().
		Str("wallet_address", event.WalletAddress).
		Str("chain", event.Chain).
		Str("event_id", event.EventID).
		Str("user_id", event.UserID).
		Str("nickname", event.Nickname).
		Msg("processing wallet tracking request")

	job := models.IngestionJob{
		Wallet: event.WalletAddress,
		Chain:  event.Chain,
	}

	if err := s.Enqueue(job); err != nil {
		s.logger.Error().
			Err(err).
			Str("wallet_address", event.WalletAddress).
			Str("chain", event.Chain).
			Msg("failed to enqueue ingestion job")
		return fmt.Errorf("enqueue job: %w", err)
	}

	s.logger.Info().
		Str("wallet_address", event.WalletAddress).
		Str("chain", event.Chain).
		Msg("wallet tracking request enqueued successfully")

	return nil
}

func (s *Service) InitConsumer(ctx context.Context) (*kafka.Consumer, error) {
	brokers := strings.Split(s.cfg.KafkaBrokers, ",")
	consumer, err := kafka.NewConsumer(
		ctx,
		brokers,
		s.cfg.KafkaConsumerGroupID,
		s.cfg.KafkaTopicWalletTracking,
		s.logger,
		s,
	)
	if err != nil {
		return nil, fmt.Errorf("create kafka consumer: %w", err)
	}
	return consumer, nil
}

func (s *Service) StartWorkers(ctx context.Context) {
	for i := 0; i < 4; i++ {
		go s.worker(ctx)
	}
}

func (s *Service) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-s.jobch:
			_ = s.handleJob(ctx, job)
		}
	}
}

func (s *Service) handleJob(ctx context.Context, job models.IngestionJob) error {
	if job.Chain == "solana" || job.Chain == "sol" {
		return s.handleSolanaJob(ctx, job)
	}

	var allTransactions []kafka.Transaction

	transfers, err := s.alc.GetAssetTransfersAll(ctx, struct {
		FromBlock    string   `json:"fromBlock,omitempty"`
		ToBlock      string   `json:"toBlock,omitempty"`
		FromAddress  string   `json:"fromAddress,omitempty"`
		ToAddress    string   `json:"toAddress,omitempty"`
		Category     []string `json:"category,omitempty"`
		WithMetadata bool     `json:"withMetadata,omitempty"`
		MaxCount     string   `json:"maxCount,omitempty"`
		PageKey      string   `json:"pageKey,omitempty"`
	}{
		FromAddress:  job.Wallet,
		MaxCount:     "0x3e8",
		WithMetadata: true,
		Category:     []string{"external", "erc20", "erc721", "erc1155"},
	})
	if err != nil {
		s.logger.Error().Err(err).Str("wallet", job.Wallet).Str("chain", job.Chain).Msg("alchemy transfers failed")
	} else {
		s.logger.Info().Int("count", len(transfers)).Str("wallet", job.Wallet).Msg("fetched alchemy transfers")
		if len(transfers) > 0 {
			if err := s.storeRaw(ctx, "alchemy", job.Wallet, job.Chain, transfers); err != nil {
				s.logger.Error().Err(err).Msg("store alchemy transfers failed")
			}

			txs, err := transformer.TransformAlchemyTransfers(transfers)
			if err != nil {
				s.logger.Error().Err(err).Msg("transform alchemy transfers failed")
			} else {
				allTransactions = append(allTransactions, txs...)
			}
		}
	}

	bal, err := s.mor.GetWalletTokenBalancesPrice(ctx, job.Wallet, job.Chain)
	if err != nil {
		s.logger.Error().Err(err).Str("wallet", job.Wallet).Msg("moralis balances failed")
	} else if bal != nil {
		if err := s.storeRaw(ctx, "moralis", job.Wallet, job.Chain, bal); err != nil {
			s.logger.Error().Err(err).Msg("store moralis balances failed")
		}
	}

	if s.producer != nil {
		event := kafka.TransactionDataIngestedEvent{
			WalletAddress: job.Wallet,
			Chain:         job.Chain,
			Transactions:  allTransactions,
		}
		if err := s.producer.PublishTransactionDataIngested(ctx, event); err != nil {
			s.logger.Error().Err(err).Str("wallet", job.Wallet).Msg("publish kafka event failed")
			return err
		}
		s.logger.Info().
			Str("wallet", job.Wallet).
			Str("chain", job.Chain).
			Int("transaction_count", len(allTransactions)).
			Msg("published kafka event successfully")
	}

	return nil
}

func (s *Service) handleSolanaJob(ctx context.Context, job models.IngestionJob) error {
	var allTransactions []kafka.Transaction

	txs, err := s.sol.GetTransactions(ctx, solana.GetTransactionsParams{
		Address: job.Wallet,
		Limit:   1000,
	})
	if err != nil {
		s.logger.Error().Err(err).Str("wallet", job.Wallet).Msg("solana transactions failed")
	} else {
		s.logger.Info().Int("count", len(txs)).Str("wallet", job.Wallet).Msg("fetched solana transactions")
		if len(txs) > 0 {
			if err := s.storeRaw(ctx, "solana", job.Wallet, job.Chain, txs); err != nil {
				s.logger.Error().Err(err).Msg("store solana transactions failed")
			}

			transformed, err := transformer.TransformSolanaTransactions(txs)
			if err != nil {
				s.logger.Error().Err(err).Msg("transform solana transactions failed")
			} else {
				allTransactions = append(allTransactions, transformed...)
			}
		}
	}

	balances, err := s.sol.GetTokenBalances(ctx, job.Wallet)
	if err != nil {
		s.logger.Error().Err(err).Str("wallet", job.Wallet).Msg("solana balances failed")
	} else if len(balances) > 0 {
		if err := s.storeRaw(ctx, "solana_balances", job.Wallet, job.Chain, balances); err != nil {
			s.logger.Error().Err(err).Msg("store solana balances failed")
		}
	}

	bal, err := s.mor.GetWalletTokenBalancesPrice(ctx, job.Wallet, job.Chain)
	if err != nil {
		s.logger.Error().Err(err).Str("wallet", job.Wallet).Msg("moralis balances failed")
	} else if bal != nil {
		if err := s.storeRaw(ctx, "moralis", job.Wallet, job.Chain, bal); err != nil {
			s.logger.Error().Err(err).Msg("store moralis balances failed")
		}
	}

	if s.producer != nil {
		event := kafka.TransactionDataIngestedEvent{
			WalletAddress: job.Wallet,
			Chain:         job.Chain,
			Transactions:  allTransactions,
		}
		if err := s.producer.PublishTransactionDataIngested(ctx, event); err != nil {
			s.logger.Error().Err(err).Str("wallet", job.Wallet).Msg("publish kafka event failed")
			return err
		}
		s.logger.Info().
			Str("wallet", job.Wallet).
			Str("chain", job.Chain).
			Int("transaction_count", len(allTransactions)).
			Msg("published kafka event successfully")
	}

	return nil
}

func (s *Service) storeRaw(ctx context.Context, source, wallet, chain string, data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal data: %w", err)
	}

	var storeErr error
	if source == "alchemy" || source == "solana" {
		_, storeErr = s.db.Pool.Exec(ctx, `
INSERT INTO raw_transactions (source_api, wallet_address, chain, data)
VALUES ($1, $2, $3, $4)
`, source, wallet, chain, b)
	} else {
		_, storeErr = s.db.Pool.Exec(ctx, `
INSERT INTO raw_balances (wallet_address, chain, data)
VALUES ($1, $2, $3)
`, wallet, chain, b)
	}

	if storeErr != nil {
		return fmt.Errorf("store raw data (source=%s): %w", source, storeErr)
	}

	return nil
}

func (s *Service) Close() error {
	s.logger.Info().Msg("closing service")
	
	if s.producer != nil {
		s.logger.Info().Msg("closing kafka producer")
		if err := s.producer.Close(); err != nil {
			return fmt.Errorf("close producer: %w", err)
		}
	}
	
	return nil
}

func (s *Service) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	status := "ok"
	kafkaStatus := "enabled"
	
	if s.cfg.KafkaEnabled {
		if s.producer == nil {
			status = "degraded"
			kafkaStatus = "unavailable"
		}
	} else {
		kafkaStatus = "disabled"
	}
	
	queueDepth := len(s.jobch)
	
	response := map[string]interface{}{
		"status":       status,
		"kafka":        kafkaStatus,
		"queue_depth":  queueDepth,
		"queue_capacity": cap(s.jobch),
	}
	
	statusCode := http.StatusOK
	if status == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}
	
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
