package repositories

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDBReservationRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoDBReservationRepository(client *dynamodb.Client, tableName string) *DynamoDBReservationRepository {
	return &DynamoDBReservationRepository{
		client:    client,
		tableName: tableName,
	}
}

// SaveReadModel writes an optimized read view of the reservation.
// PK: EVENT#<event_id>
// SK: RES#<start_time>#<reservation_id>
func (r *DynamoDBReservationRepository) SaveReadModel(ctx context.Context, reservationID, eventID, startTime, status string) error {
	pk := fmt.Sprintf("EVENT#%s", eventID)
	// ISO8601 strings sort lexicographically
	sk := fmt.Sprintf("RES#%s#%s", startTime, reservationID)

	_, err := r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item: map[string]types.AttributeValue{
			"PK":            &types.AttributeValueMemberS{Value: pk},
			"SK":            &types.AttributeValueMemberS{Value: sk},
			"ReservationID": &types.AttributeValueMemberS{Value: reservationID},
			"Status":        &types.AttributeValueMemberS{Value: status},
			"UpdatedAt":     &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
	})

	if err != nil {
		log.Printf("Failed to write to DynamoDB: %v", err)
		return err
	}
	return nil
}
