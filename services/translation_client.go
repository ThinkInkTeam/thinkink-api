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
	Msk []float32   `json:"msk"`
}

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

// TranslateDummyData sends dummy EEG data for testing purposes
func (tc *TranslationClient) TranslateDummyData(token string) ([]string, error) {
	// Create dummy EEG data for testing (56 timesteps, 840 features each)
	dummyEeg := make([][]float32, 56)
	for i := range dummyEeg {
		dummyEeg[i] = make([]float32, 840)
		// Fill with zeros or random values for testing
		for j := range dummyEeg[i] {
			dummyEeg[i][j] = 0.0
		}
	}
	
	// Create dummy mask data
	dummyMsk := make([]float32, 56)
	for i := range dummyMsk {
		dummyMsk[i] = 1.0 // All ones for testing
	}

	return tc.TranslateEEG(token, dummyEeg, dummyMsk)
}

// MarshalJSON customizes the JSON encoding for EEGData
func (eegData *EEGData) MarshalJSON() ([]byte, error) {
	type Alias EEGData
	return json.Marshal(&struct {
		Eeg [][]float32 `json:"eeg"`
		Msk []float32   `json:"msk"`
		*Alias
	}{
		Eeg:   eegData.Eeg,
		Msk:   eegData.Msk,
		Alias: (*Alias)(eegData),
	})
}

// UnmarshalJSON customizes the JSON decoding for EEGData
func (eegData *EEGData) UnmarshalJSON(data []byte) error {
	type Alias EEGData
	aux := &struct {
		Eeg [][]float32 `json:"eeg"`
		Msk []float32   `json:"msk"`
		*Alias
	}{
		Alias: (*Alias)(eegData),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	eegData.Eeg = aux.Eeg
	eegData.Msk = aux.Msk
	return nil
}

// ParseEEGData parses byte data into structured EEG format
func ParseEEGData(data []byte) ([][]float32, []float32, error) {
	var eegData EEGData
	
	// Try to parse as JSON first
	if err := json.Unmarshal(data, &eegData); err == nil {
		return eegData.Eeg, eegData.Msk, nil
	}
	
	// If JSON parsing fails, create dummy data for now
	// This is a fallback until the proper file format is established
	log.Printf("Warning: Could not parse EEG data as JSON, using dummy data")
	
	// Create dummy EEG data (56 timesteps, 840 features each)
	dummyEeg := make([][]float32, 56)
	for i := range dummyEeg {
		dummyEeg[i] = make([]float32, 840)
		for j := range dummyEeg[i] {
			dummyEeg[i][j] = 0.0
		}
	}
	
	// Create dummy mask data
	dummyMsk := make([]float32, 56)
	for i := range dummyMsk {
		dummyMsk[i] = 1.0
	}
	
	return dummyEeg, dummyMsk, nil
}

// TranslateEEGFromBytes parses byte data and sends it to the ML server for translation
func (tc *TranslationClient) TranslateEEGFromBytes(token string, data []byte) ([]string, error) {
	eeg, msk, err := ParseEEGData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EEG data: %v", err)
	}
	
	return tc.TranslateEEG(token, eeg, msk)
}
