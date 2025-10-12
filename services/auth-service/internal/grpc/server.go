// Package grpcserver implements the gRPC server for authentication services.
package grpcserver

import (
	"context"

	gen "github.com/aezizhu/million-dollar-hunter/services/auth-service/api/gen"
	jwtmgr "github.com/aezizhu/million-dollar-hunter/services/auth-service/internal/jwt"
)

type Server struct {
	gen.UnimplementedAuthServiceServer
	JWT *jwtmgr.Manager
}

func New(jwt *jwtmgr.Manager) *Server {
	return &Server{JWT: jwt}
}

func (s *Server) GenerateTokens(ctx context.Context, req *gen.TokenRequest) (*gen.TokenPair, error) {
	access, refresh, exp, err := s.JWT.GeneratePair(req.GetUserId(), req.GetEmail())
	if err != nil {
		return nil, err
	}
	return &gen.TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    exp.Unix(),
	}, nil
}

// ValidateToken validates a JWT token. Context is currently not used to abort validation since the operation is synchronous and CPU-bound.
// The context is kept for future extensibility (e.g., revocation checks against storage). Canceled contexts do not affect the result.
func (s *Server) ValidateToken(ctx context.Context, req *gen.ValidateRequest) (*gen.ValidateResponse, error) {
	claims, err := s.JWT.ValidateToken(req.GetToken(), req.GetExpectedAud())
	if err != nil {
		return &gen.ValidateResponse{
			Valid:  false,
			Reason: err.Error(),
		}, nil
	}
	return &gen.ValidateResponse{
		Valid:  true,
		UserId: claims.Subject,
		Email:  claims.Email,
	}, nil
}
