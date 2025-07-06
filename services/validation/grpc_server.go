package validation

import (
	"context"
	"log"

	pb "github.com/ThinkInkTeam/thinkink-core-backend/proto-gen/proto/validation"
)

// Server implements the TokenValidationService gRPC server
type Server struct {
	pb.UnimplementedTokenValidationServiceServer
	tokenValidator *TokenValidator
}

// NewServer creates a new gRPC validation server
func NewServer() *Server {
	return &Server{
		tokenValidator: NewTokenValidator(),
	}
}

// ValidateMLToken implements the gRPC service method
func (s *Server) ValidateMLToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	log.Printf("Validating token for ML service: %s", req.Token[:10]+"...") // Log only first 10 chars for security

	isValid := s.tokenValidator.ValidateToken(req.Token)

	log.Printf("Token validation result: %v", isValid)

	return &pb.ValidateTokenResponse{
		IsValid: isValid,
	}, nil
}
