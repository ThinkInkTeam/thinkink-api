package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	translationpb "github.com/ThinkInkTeam/thinkink-core-backend/proto-gen/proto/translation"
)

// TranslationClient wraps the gRPC translation client
type TranslationClient struct {
	conn   *grpc.ClientConn
	client translationpb.TranslationServiceClient
}

// NewTranslationClient creates a new translation client
func NewTranslationClient(address string) (*TranslationClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to translation service: %v", err)
	}

	client := translationpb.NewTranslationServiceClient(conn)
	
	return &TranslationClient{
		conn:   conn,
		client: client,
	}, nil
}

// Close closes the gRPC connection
func (tc *TranslationClient) Close() error {
	return tc.conn.Close()
}

// TranslateEEG sends EEG data to the ML server for translation
func (tc *TranslationClient) TranslateEEG(token string, eegData []byte) ([]string, error) {
	// Clean token (remove Bearer prefix if present)
	cleanToken := strings.TrimPrefix(strings.TrimSpace(token), "Bearer ")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create the request
	req := &translationpb.TranslateRequest{
		Token:   cleanToken,
		Content: string(eegData),
	}

	// Call the translation service
	log.Printf("Sending translation request to ML server")
	resp, err := tc.client.Translate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("translation request failed: %v", err)
	}

	// Check for errors in response
	if resp.ErrorMessage != "" {
		return nil, fmt.Errorf("translation error: %s", resp.ErrorMessage)
	}

	log.Printf("Translation successful: %v", resp.Translated)
	return resp.Translated, nil
}

// TranslateDummyData sends dummy EEG data for testing purposes
func (tc *TranslationClient) TranslateDummyData(token string) ([]string, error) {
	// Create dummy EEG data for testing (56 timesteps, 840 features each)
	dummyData := []byte(strings.Repeat("0.0,", 56*840))


	return tc.TranslateEEG(token, dummyData)
}
