package services

import (
	"encoding/json"
	"fmt"

	"github.com/DulsaraNethmin/shopware-shopify-integration/internal/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sfn"
	"gorm.io/gorm"
)

// StepFunctionsService handles AWS Step Functions operations
type StepFunctionsService struct {
	config config.AWSConfig
	db     *gorm.DB
	client *sfn.SFN
}

// NewStepFunctionsService creates a new Step Functions service
func NewStepFunctionsService(config config.AWSConfig, db *gorm.DB) *StepFunctionsService {
	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(config.Region),
		Credentials: credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, ""),
	})

	if err != nil {
		// Log error but continue - we'll check for client before using
		fmt.Printf("Error creating AWS session: %v\n", err)
	}

	var client *sfn.SFN
	if sess != nil {
		client = sfn.New(sess)
	}

	return &StepFunctionsService{
		config: config,
		db:     db,
		client: client,
	}
}

// MigrationInput represents the input to the Step Functions state machine
type MigrationInput struct {
	DataflowID  uint            `json:"dataflow_id"`
	MigrationID uint            `json:"migration_id"`
	SourceData  json.RawMessage `json:"source_data"`
}

// StartExecution starts a Step Functions execution
func (s *StepFunctionsService) StartExecution(dataflowID, migrationID uint, sourceData json.RawMessage) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("AWS Step Functions client not initialized")
	}

	// Create execution input
	input := MigrationInput{
		DataflowID:  dataflowID,
		MigrationID: migrationID,
		SourceData:  sourceData,
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("error marshaling execution input: %w", err)
	}

	// Start execution
	result, err := s.client.StartExecution(&sfn.StartExecutionInput{
		StateMachineArn: aws.String(s.config.StepFunctionsARN),
		Input:           aws.String(string(inputJSON)),
	})

	if err != nil {
		return "", fmt.Errorf("error starting Step Functions execution: %w", err)
	}

	return *result.ExecutionArn, nil
}

// GetExecutionStatus gets the status of a Step Functions execution
func (s *StepFunctionsService) GetExecutionStatus(executionARN string) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("AWS Step Functions client not initialized")
	}

	result, err := s.client.DescribeExecution(&sfn.DescribeExecutionInput{
		ExecutionArn: aws.String(executionARN),
	})

	if err != nil {
		return "", fmt.Errorf("error describing Step Functions execution: %w", err)
	}

	return *result.Status, nil
}

// GetExecutionResults gets the results of a Step Functions execution
func (s *StepFunctionsService) GetExecutionResults(executionARN string) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("AWS Step Functions client not initialized")
	}

	result, err := s.client.DescribeExecution(&sfn.DescribeExecutionInput{
		ExecutionArn: aws.String(executionARN),
	})

	if err != nil {
		return "", fmt.Errorf("error describing Step Functions execution: %w", err)
	}

	if result.Output == nil {
		return "", nil
	}

	return *result.Output, nil
}
