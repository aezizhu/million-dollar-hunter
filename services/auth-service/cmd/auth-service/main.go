package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"google.golang.org/grpc"

	gen "github.com/aezizhu/million-dollar-hunter/services/auth-service/api/gen"
	"github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/config"
	grpcserver "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/grpc"
	httpapi "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/http"
	jwtmgr "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/jwt"
	secrets "github.com/aezizhu/million-dollar-hunter/pkg/secrets"
	"github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/store"
)

type jwtSecret struct {
	KID string `json:"kid"`
	Key string `json:"key"`
}

func main() {
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := config.Parse()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse config")
	}

	multiUser := false
	if val := os.Getenv("ENABLE_MULTI_USER"); val != "" {
		switch strings.ToLower(val) {
		case "true", "1", "yes":
			multiUser = true
		case "false", "0", "no":
			multiUser = false
		default:
			log.Fatal().Str("value", val).Msg("ENABLE_MULTI_USER must be true/false/1/0/yes/no")
		}
	}

	if !multiUser {
		if os.Getenv("MVP_USERNAME") == "" {
			log.Fatal().Msg("MVP_USERNAME is required in MVP mode")
		}
		mvpPass := os.Getenv("MVP_PASSWORD")
		if mvpPass == "" {
			log.Fatal().Msg("MVP_PASSWORD is required in MVP mode")
		}
		if len(mvpPass) != 60 || (!strings.HasPrefix(mvpPass, "$2a$") && !strings.HasPrefix(mvpPass, "$2b$") && !strings.HasPrefix(mvpPass, "$2y$")) {
			log.Warn().Int("length", len(mvpPass)).
				Msg("MVP_PASSWORD may not be a valid bcrypt hash (expected 60 chars starting with $2a/$2b/$2y)")
		}
		log.Info().Msg("Running in MVP mode")
	} else {
		if os.Getenv("DATABASE_URL") == "" {
			log.Fatal().Msg("DATABASE_URL is required in multi-user mode")
		}
		log.Info().Msg("Running in multi-user mode")
	}

	var secClient secrets.Client
	switch strings.ToLower(os.Getenv("SECRETS_PROVIDER")) {
	case "aws":
		region := os.Getenv("AWS_REGION")
		awsClient, awsErr := secrets.NewAWS(context.Background(), secrets.AWSConfig{
			Config: secrets.Config{
				CacheTTL:        time.Hour,
				RefreshInterval: time.Minute,
			},
			Region: region,
		})
		if awsErr != nil {
			log.Error().Err(awsErr).Msg("failed to init AWS secrets client, falling back to env")
			secClient = secrets.NewEnv(secrets.Config{CacheTTL: time.Hour, RefreshInterval: time.Minute})
		} else {
			secClient = awsClient
		}
	default:
		secClient = secrets.NewEnv(secrets.Config{CacheTTL: time.Hour, RefreshInterval: time.Minute})
	}
	if secClient != nil {
		secClient.StartBackgroundRefresh(context.Background())
	}

	keys := cfg.JWTKeys
	currentKID := cfg.JWTCurrentKID

	secretPrefix := os.Getenv("SECRETS_PREFIX")
	if secretPrefix == "" {
		env := os.Getenv("ENV")
		if env == "" {
			env = "dev"
		}
		secretPrefix = "mdh/" + env + "/auth/jwt"
	}
	if secClient != nil && strings.ToLower(os.Getenv("SECRETS_PROVIDER")) != "" {
		var cur jwtSecret
		if getErr := secClient.GetJSON(context.Background(), secretPrefix+"/current", &cur); getErr == nil && cur.KID != "" && cur.Key != "" {
			if keys == nil {
				keys = map[string][]byte{}
			}
			keys[cur.KID] = []byte(cur.Key)
			currentKID = cur.KID
		}
		var prev jwtSecret
		if getErr := secClient.GetJSON(context.Background(), secretPrefix+"/previous", &prev); getErr == nil && prev.KID != "" && prev.Key != "" {
			if keys == nil {
				keys = map[string][]byte{}
			}
			keys[prev.KID] = []byte(prev.Key)
		}
	}

	var j interface {
		GeneratePair(userID, email string) (string, string, time.Time, error)
		ValidateToken(tokenStr string, expectedAud string) (*jwtmgr.Claims, error)
	}

	if keystorePath := os.Getenv("KEYSTORE_PATH"); keystorePath != "" {
		ks, kerr := jwtmgr.NewKeyStore(keystorePath)
		if kerr != nil {
			log.Fatal().Err(kerr).Str("path", keystorePath).Msg("failed to load keystore")
		}
		j = jwtmgr.NewWithKeyStore(cfg.JWTIssuer, cfg.JWTAudience, cfg.AccessTTL, cfg.RefreshTTL, ks)
		log.Info().Str("keystore", keystorePath).Msg("Using RS256 keystore mode")
	} else {
		var mgr *jwtmgr.Manager
		if currentKID != "" && len(keys) > 0 {
			mgr = jwtmgr.NewWithKeys(cfg.JWTIssuer, cfg.JWTAudience, cfg.AccessTTL, cfg.RefreshTTL, keys, currentKID)
		} else {
			mgr = jwtmgr.New(cfg.JWTIssuer, cfg.JWTAudience, cfg.AccessTTL, cfg.RefreshTTL, cfg.JWTSigningKey)
		}
		j = mgr

		if secClient != nil && strings.ToLower(os.Getenv("SECRETS_PROVIDER")) != "" {
			go func() {
				t := time.NewTicker(1 * time.Minute)
				defer t.Stop()
				for range t.C {
					var cur jwtSecret
					if getErr := secClient.GetJSON(context.Background(), secretPrefix+"/current", &cur); getErr == nil && cur.KID != "" && cur.Key != "" {
						nk := map[string][]byte{cur.KID: []byte(cur.Key)}
						var prev jwtSecret
						if getErr := secClient.GetJSON(context.Background(), secretPrefix+"/previous", &prev); getErr == nil && prev.KID != "" && prev.Key != "" {
							nk[prev.KID] = []byte(prev.Key)
						}
						mgr.UpdateKeys(nk, cur.KID)
					}
				}
			}()
		}
	}

	mux := http.NewServeMux()
	s := &httpapi.Server{Logger: &log, JWT: j}
	if multiUser {
		dsn := os.Getenv("DATABASE_URL")
		pool, poolErr := pgxpool.New(context.Background(), dsn)
		if poolErr != nil {
			log.Fatal().Err(poolErr).Msg("database connection failed")
		}
		pg := &store.PGStore{Pool: pool}
		s.Store = pg
		s.RefreshTokens = pg
		s.Audit = pg
		defer pool.Close()
	}
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		type health struct {
			OK            bool   `json:"ok"`
			SecretsStatus string `json:"secrets"`
		}
		h := health{OK: true, SecretsStatus: "disabled"}
		if secClient != nil && strings.ToLower(os.Getenv("SECRETS_PROVIDER")) != "" {
			if healthErr := secClient.Health(r.Context()); healthErr != nil {
				h.SecretsStatus = "degraded"
			} else {
				h.SecretsStatus = "ok"
			}
		}
		_ = json.NewEncoder(w).Encode(h)
	})
	mux.HandleFunc("/api/v1/auth/login", s.Login)
	mux.HandleFunc("/api/v1/auth/register", s.Register)
	mux.HandleFunc("/api/v1/auth/logout", s.Logout)
	mux.HandleFunc("/api/v1/auth/refresh", s.Refresh)
	mux.HandleFunc("/.well-known/jwks.json", s.JWKS)

	httpSrv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           hlog.NewHandler(log)(mux),
		ReadHeaderTimeout: 30 * time.Second,
	}

	grpcSrv := grpc.NewServer()
	gen.RegisterAuthServiceServer(grpcSrv, grpcserver.New(j))

	grpcLn, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Error().Err(err).Msg("failed to start gRPC listener")
		return
	}

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()
	go func() {
		if err := grpcSrv.Serve(grpcLn); err != nil {
			log.Fatal().Err(err).Msg("gRPC server failed")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(ctx)
	grpcSrv.GracefulStop()
}
