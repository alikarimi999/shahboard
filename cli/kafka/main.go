package main

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/alikarimi999/shahboard/types"
	"github.com/spf13/cobra"
)

func main() {
	var brokerAddress, groupID string

	var rootCmd = &cobra.Command{
		Use:   "cli-tool",
		Short: "CLI tool for sending events to Kafka",
	}

	rootCmd.PersistentFlags().StringVar(&brokerAddress, "broker", "localhost:9092", "Kafka broker address")
	rootCmd.PersistentFlags().StringVar(&groupID, "group", types.NewObjectId().String(), "Kafka group ID")

	rootCmd.AddCommand(listenCommand(newConsumerGroup(brokerAddress, groupID)))

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
	}
}

func newConsumerGroup(brokerAddress, groupID string) sarama.ConsumerGroup {
	c, err := sarama.NewConsumerGroup([]string{brokerAddress}, groupID, sarama.NewConfig())
	if err != nil {
		panic(err)
	}
	return c
}

func listenCommand(c sarama.ConsumerGroup) *cobra.Command {
	var domain, action string

	cmd := &cobra.Command{
		Use:   "listen",
		Short: "Listen to events",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Listening to events...")
			return c.Consume(context.Background(), []string{domain}, &handler{action: action})
		},
	}

	cmd.Flags().StringVar(&domain, "domain", "", "Event domain")
	cmd.Flags().StringVar(&action, "action", "", "Event action")

	return cmd
}

type handler struct {
	action string
}

func (h *handler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *handler) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (h *handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		for _, header := range msg.Headers {
			if h.action != "" {
				if string(header.Key) == "action" && string(header.Value) == h.action {
					fmt.Println(msg.Offset, string(msg.Value))
					continue
				}
				continue
			}
			fmt.Println(msg.Offset, string(msg.Value))
		}
	}
	return nil
}
