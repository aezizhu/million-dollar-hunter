package service

import (
	"context"
	"encoding/json"

	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/alchemy"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/breaker"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/config"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/models"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/moralis"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/ratelimit"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/repository"
	"github.com/aezizhu/million-dollar-hunter/ingestion-service/internal/solana"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
)

type Service struct {
	cfg     *config.Config
	logger  zerolog.Logger
	db      *repository.Postgres
	alc     *alchemy.Client
	mor     *moralis.Client
	sol     *solana.Client
	jobch   chan models.IngestionJob
}

func New(ctx context.Context, cfg *config.Config, logger zerolog.Logger, db *repository.Postgres) (*Service, error) {
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
	return &Service{
		cfg:    cfg,
		logger: logger,
		db:     db,
		alc:    alc,
		mor:    mor,
		sol:    sol,
		jobch:  make(chan models.IngestionJob, 64),
	}, nil
}

func (s *Service) Enqueue(job models.IngestionJob) {
	select {
	case s.jobch <- job:
	default:
	}
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
	}

	if len(transfers) > 0 {
		if err := s.storeRaw(ctx, "alchemy", job.Wallet, job.Chain, transfers); err != nil {
			s.logger.Error().Err(err).Msg("store transfers")
		}
	}

	bal, err := s.mor.GetWalletTokenBalancesPrice(ctx, job.Wallet, job.Chain)
	if err == nil {
		if err := s.storeRaw(ctx, "moralis", job.Wallet, job.Chain, bal); err != nil {
			s.logger.Error().Err(err).Msg("store balances")
		}
	}

	return nil
}

func (s *Service) handleSolanaJob(ctx context.Context, job models.IngestionJob) error {
	txs, err := s.sol.GetTransactions(ctx, solana.GetTransactionsParams{
		Address: job.Wallet,
		Limit:   1000,
	})
	if err != nil {
		s.logger.Error().Err(err).Str("wallet", job.Wallet).Msg("solana transactions failed")
	}

	if len(txs) > 0 {
		if err := s.storeRaw(ctx, "solana", job.Wallet, job.Chain, txs); err != nil {
			s.logger.Error().Err(err).Msg("store solana transactions")
		}
	}

	balances, err := s.sol.GetTokenBalances(ctx, job.Wallet)
	if err != nil {
		s.logger.Error().Err(err).Str("wallet", job.Wallet).Msg("solana balances failed")
	}

	if len(balances) > 0 {
		if err := s.storeRaw(ctx, "solana_balances", job.Wallet, job.Chain, balances); err != nil {
			s.logger.Error().Err(err).Msg("store solana balances")
		}
	}

	bal, err := s.mor.GetWalletTokenBalancesPrice(ctx, job.Wallet, job.Chain)
	if err == nil {
		if err := s.storeRaw(ctx, "moralis", job.Wallet, job.Chain, bal); err != nil {
			s.logger.Error().Err(err).Msg("store moralis balances")
		}
	}

	return nil
}

func (s *Service) storeRaw(ctx context.Context, source, wallet, chain string, data any) error {
	b, _ := json.Marshal(data)
	_, err := s.db.Pool.Exec(ctx, `
INSERT INTO raw_transactions (source_api, wallet_address, chain, data)
SELECT $1, $2, $3, $4
`, source, wallet, chain, b)
	if err != nil {
		_, err2 := s.db.Pool.Exec(ctx, `
INSERT INTO raw_balances (wallet_address, chain, data)
SELECT $1, $2, $3
`, wallet, chain, b)
		return err2
	}
	return nil
}
