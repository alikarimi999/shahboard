package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/types"
	"github.com/spf13/cobra"
)

func main() {
	var brokerAddress string

	// Define the root command
	var rootCmd = &cobra.Command{
		Use:   "cli-tool",
		Short: "CLI tool for sending events to Kafka",
	}

	// Add global flag for broker address
	rootCmd.PersistentFlags().StringVar(&brokerAddress, "broker", "localhost:9092", "Kafka broker address")

	// Add subcommands
	rootCmd.AddCommand(newPlayersMatchedCommand(&brokerAddress))
	rootCmd.AddCommand(newPlayerMovedCommand(&brokerAddress))

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
	}
}

func newPlayersMatchedCommand(brokerAddress *string) *cobra.Command {
	var player1, player2 int64

	cmd := &cobra.Command{
		Use:   "players_matched",
		Short: "Send a PlayersMatched event",
		RunE: func(cmd *cobra.Command, args []string) error {
			event := event.EventPlayersMatched{
				Player1:   types.ObjectId(player1),
				Player2:   types.ObjectId(player2),
				Timestamp: time.Now().Unix(),
			}
			return sendEventToKafka(*brokerAddress, event)
		},
	}

	// Add specific flags for this command
	cmd.Flags().Int64Var(&player1, "player1", int64(types.NewObjectId()), "Player1 ID")
	cmd.Flags().Int64Var(&player2, "player2", int64(types.NewObjectId()), "Player2 ID")

	return cmd
}

func newPlayerMovedCommand(brokerAddress *string) *cobra.Command {
	var gameId, playerId int64
	var move string

	cmd := &cobra.Command{
		Use:   "player_moved",
		Short: "Send a GamePlayerMoved event",
		RunE: func(cmd *cobra.Command, args []string) error {
			event := event.EventGamePlayerMoved{
				ID:        types.ObjectId(gameId),
				PlayerID:  types.ObjectId(playerId),
				Move:      move,
				Timestamp: time.Now().Unix(),
			}
			return sendEventToKafka(*brokerAddress, event)
		},
	}

	// Add specific flags for this command
	cmd.Flags().Int64Var(&gameId, "gameId", int64(types.NewObjectId()), "Game ID")
	cmd.Flags().Int64Var(&playerId, "playerId", int64(types.NewObjectId()), "Player ID")
	cmd.Flags().StringVar(&move, "move", "", "Move")

	cmd.MarkFlagRequired("gameId")
	cmd.MarkFlagRequired("playerId")
	cmd.MarkFlagRequired("move")

	return cmd
}

func sendEventToKafka(brokerAddress string, event event.Event) error {
	// Configure Sarama producer
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll       // Wait for all in-sync replicas to acknowledge
	config.Producer.Retry.Max = 5                          // Retry up to 5 times
	config.Producer.Return.Successes = true                // Return success messages
	config.Producer.Compression = sarama.CompressionSnappy // Use Snappy compression
	config.Version = sarama.V2_5_0_0                       // Specify Kafka version

	// Create a Sarama producer
	producer, err := sarama.NewSyncProducer([]string{brokerAddress}, config)
	if err != nil {
		return fmt.Errorf("failed to create Kafka producer: %w", err)
	}
	defer func() {
		if err := producer.Close(); err != nil {
			log.Printf("failed to close Kafka producer: %v\n", err)
		}
	}()

	t := event.GetTopic().Domain().String()
	// Serialize the event to JSON
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create a Sarama message
	message := &sarama.ProducerMessage{
		Topic: t, // Use the event's topic
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("action"),
				Value: []byte(event.GetAction().String()),
			},
		},
		Key:   sarama.ByteEncoder(event.GetTopic().String()),
		Value: sarama.ByteEncoder(eventBytes),
	}

	// Send the message to Kafka
	partition, offset, err := producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	fmt.Printf("Event sent to topic: %s successfully! Partition: %d, Offset: %d\n", t, partition, offset)
	return nil
}

// go run cli/game/kafka/main.go players_matched

// gameID: 7459746929626178861  w:7459746931359953653 b:7459746933042382070

// go run cli/game/kafka/main.go player_moved --gameId 7459746929626178861 --playerId 7459746931359953653 --move Na3
