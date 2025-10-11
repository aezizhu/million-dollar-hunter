# Auth Service

gRPC-based authentication service providing JWT generation and validation for internal services.

- Port: configurable via server main; defaults follow service config
- RPCs:
  - GenerateTokens(user_id, email) -> TokenPair { access_token, refresh_token, expires_in }
  - ValidateToken(token, expected_aud) -> ValidateResponse { valid, user_id, email, reason }

Current ValidateToken semantics:
- Invalid tokens return a successful gRPC response with valid=false and a populated reason
- Transport errors (e.g., connection issues, deadlines) return gRPC errors
- Gateway treats valid=false and client/transport errors as unauthorized unless fallback is enabled

Sample Go client:

package main

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	gen "github.com/aezizhu/million-dollar-hunter/services/auth-service/api/gen"
)

func main() {
	conn, err := grpc.Dial("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	cli := gen.NewAuthServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := cli.ValidateToken(ctx, &gen.ValidateRequest{
		Token:       "access.jwt.here",
		ExpectedAud: "aud",
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("valid=%v user_id=%s reason=%s\n", resp.GetValid(), resp.GetUserId(), resp.GetReason())
}
