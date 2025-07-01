package validation

import (
	"context"
	"log"
)

// Server implements the TokenValidationService gRPC server
type Server struct {
	UnimplementedTokenValidationServiceServer
	tokenValidator *TokenValidator
}

// NewServer creates a new gRPC validation server
func NewServer() *Server {
	return &Server{
		tokenValidator: NewTokenValidator(),
	}
}

// ValidateMLToken implements the gRPC service method
func (s *Server) ValidateMLToken(ctx context.Context, req *ValidateTokenRequest) (*ValidateTokenResponse, error) {
	log.Printf("Validating token for ML service: %s", req.Token[:10]+"...") // Log only first 10 chars for security
	
	isValid := s.tokenValidator.ValidateToken(req.Token)
	
	log.Printf("Token validation result: %v", isValid)
	
	return &ValidateTokenResponse{
		IsValid: isValid,
	}, nil
}
