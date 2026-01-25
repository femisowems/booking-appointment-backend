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

type DynamoDBAppointmentRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoDBAppointmentRepository(client *dynamodb.Client, tableName string) *DynamoDBAppointmentRepository {
	return &DynamoDBAppointmentRepository{
		client:    client,
		tableName: tableName,
	}
}

// SaveReadModel writes an optimized read view of the appointment.
// PK: PROVIDER#<provider_id>
// SK: APPT#<start_time>#<appt_id>
func (r *DynamoDBAppointmentRepository) SaveReadModel(ctx context.Context, apptID, providerID, startTime, status string) error {
	pk := fmt.Sprintf("PROVIDER#%s", providerID)
	// ISO8601 strings sort lexicographically
	sk := fmt.Sprintf("APPT#%s#%s", startTime, apptID)

	_, err := r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item: map[string]types.AttributeValue{
			"PK":            &types.AttributeValueMemberS{Value: pk},
			"SK":            &types.AttributeValueMemberS{Value: sk},
			"AppointmentID": &types.AttributeValueMemberS{Value: apptID},
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
