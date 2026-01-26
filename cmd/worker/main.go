package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/femisowemimo/booking-appointment/backend/pkg/adapters/messaging"
	"github.com/femisowemimo/booking-appointment/backend/pkg/adapters/repositories"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it (using system env)")
	}

	log.Println("Starting Reservation Worker Service...")

	// 1. Initialize RabbitMQ
	amqpConnStr := os.Getenv("RABBITMQ_URL")
	if amqpConnStr == "" {
		amqpConnStr = "amqp://user:password@localhost:5672/"
	}
	rabbitConn, err := amqp.Dial(amqpConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitConn.Close()

	// 2. Initialize DynamoDB Client (LocalStack compatible)
	// Force custom resolver for LocalStack if env var present
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				localstackURL := os.Getenv("AWS_ENDPOINT_URL")
				if localstackURL == "" {
					localstackURL = "http://localhost:4566"
				}
				return aws.Endpoint{
					URL:           localstackURL,
					SigningRegion: "us-east-1",
				}, nil
			}),
		),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	dynamoClient := dynamodb.NewFromConfig(cfg)

	// Create table if not exists (for local dev convenience)
	ensureTableExists(context.TODO(), dynamoClient, "ReservationsReadModel")

	repo := repositories.NewDynamoDBReservationRepository(dynamoClient, "ReservationsReadModel")

	// 3. Start Worker
	worker := messaging.NewWorker(rabbitConn, repo)
	log.Fatalf("Worker exited: %v", worker.Start())
}

func ensureTableExists(ctx context.Context, client *dynamodb.Client, tableName string) {
	_, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})

	if err == nil {
		log.Printf("Table %s already exists", tableName)
		return
	}

	log.Printf("Table %s does not exist, creating...", tableName)
	_, err = client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("PK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("SK"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("PK"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("SK"), KeyType: types.KeyTypeRange},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	})

	if err != nil {
		log.Printf("Failed to create table: %v", err)
	} else {
		log.Printf("Table %s created successfully", tableName)
	}
}
