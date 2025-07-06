package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	translationpb "github.com/ThinkInkTeam/thinkink-core-backend/proto-gen/proto/translation"
)

// EEGData represents the structure expected for EEG data
type EEGData struct {
	Eeg [][]float32 `json:"eeg"`
	Msk []float32   `json:"mask"`
}

// TranslationClient wraps the gRPC translation client
type TranslationClient struct {
	conn   *grpc.ClientConn
	client translationpb.TranslationServiceClient
}

// NewTranslationClient creates a new translation client with retry logic
func NewTranslationClient(address string) (*TranslationClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), // Wait for connection to be ready
		grpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to translation service at %s: %v", address, err)
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
func (tc *TranslationClient) TranslateEEG(token string, eeg [][]float32, msk []float32) ([]string, error) {
	// Clean token (remove Bearer prefix if present)
	cleanToken := strings.TrimPrefix(strings.TrimSpace(token), "Bearer ")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Convert 2D EEG data to protobuf format
	eegRows := make([]*translationpb.EegRow, len(eeg))
	for i, row := range eeg {
		eegRows[i] = &translationpb.EegRow{Values: row}
	}

	// Create the request
	req := &translationpb.TranslateRequest{
		Token: cleanToken,
		Eeg:   eegRows,
		Msk:   msk,
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

// ParseEEGData parses byte data into structured EEG format
func ParseEEGData(data []byte) ([][]float32, []float32, error) {
	var eegData EEGData

	err := json.Unmarshal(data, &eegData)
	// Try to parse as JSON first
	if err != nil {
		return nil, nil, fmt.Errorf("err: %e", err)
	}
	return eegData.Eeg, eegData.Msk, nil

}

// TranslateEEGFromBytes parses byte data and sends it to the ML server for translation
func (tc *TranslationClient) TranslateEEGFromBytes(token string, data []byte) ([]string, error) {
	eeg, msk, err := ParseEEGData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EEG data: %v", err)
	}

	return tc.TranslateEEG(token, eeg, msk)
}
