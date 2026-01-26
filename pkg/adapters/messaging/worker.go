package messaging

import (
	"context"
	"encoding/json"
	"log"

	"github.com/femisowemimo/booking-appointment/backend/pkg/adapters/repositories"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Worker struct {
	conn       *amqp.Connection
	dynamoRepo *repositories.DynamoDBReservationRepository
}

func NewWorker(conn *amqp.Connection, dynamoRepo *repositories.DynamoDBReservationRepository) *Worker {
	return &Worker{
		conn:       conn,
		dynamoRepo: dynamoRepo,
	}
}

func (w *Worker) Start() error {
	ch, err := w.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// Ensure Queue Exists
	q, err := ch.QueueDeclare(
		"reservation_updates", // name
		true,                  // durable
		false,                 // delete when unused
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		return err
	}

	// Bind Queue to Exchange
	err = ch.QueueBind(
		q.Name,
		"reservation.#",   // routing key (was appointment.#)
		"events_exchange", // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer tag
		false,  // auto-ack (We want manual ack for reliability)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	done := make(chan struct{})

	go func() {
		defer close(done)
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			if err := w.processMessage(d.Body); err != nil {
				log.Printf("Error processing message: %v", err)
				// Basic retry strategy: Nack with requeue (dangerous loop if permanent fail)
				// For prod, use DLQ or retry count.
				d.Nack(false, true)
			} else {
				d.Ack(false)
			}
		}
		log.Println("Consumer loop exited")
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-done

	return nil
}

type EventPayload struct {
	EventType     string `json:"event_type"`
	ReservationID string `json:"reservation_id"`
	EventID       string `json:"event_id"`
	StartTime     string `json:"start_time"` // Simplified: string in JSON
	Status        string `json:"status"`     // Inferred or passed
}

func (w *Worker) processMessage(body []byte) error {
	var event EventPayload
	if err := json.Unmarshal(body, &event); err != nil {
		return err
	}

	// Update Read Model
	// Assuming raw timestamp string from JSON (RFC3339)
	status := "BOOKED" // Default for creation event
	if event.EventType == "ReservationCancelled" {
		status = "CANCELLED"
	} else if event.Status != "" {
		status = event.Status
	}

	// Make idempotent write to DynamoDB
	return w.dynamoRepo.SaveReadModel(context.Background(), event.ReservationID, event.EventID, event.StartTime, status)
}
