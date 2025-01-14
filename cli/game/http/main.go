package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/alikarimi999/shahboard/types"
	"github.com/spf13/cobra"
)

func main() {
	var httpAddress string

	// Define the root command
	var rootCmd = &cobra.Command{
		Use:   "cli-tool",
		Short: "CLI tool for sending requests to HTTP server",
	}

	rootCmd.PersistentFlags().StringVar(&httpAddress, "http", "localhost:8081", "HTTP server address")

	// Add subcommands
	rootCmd.AddCommand(getLiveGames(httpAddress))
	rootCmd.AddCommand(getGamesFEN(httpAddress))
	rootCmd.AddCommand(getGamePGN(httpAddress))

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
	}
}

func getLiveGames(httpAddress string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "live_games",
		Short: "Send a GetLiveGames request",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/games/live",
				httpAddress), nil)
			if err != nil {
				return err
			}

			return sendHttpRequest(httpAddress, req)
		}}

	return cmd
}

func getGamesFEN(httpAddress string) *cobra.Command {
	var games []int64

	cmd := &cobra.Command{
		Use:   "games_fen",
		Short: "Send a GetGamesFEN request",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(games) == 0 {
				return fmt.Errorf("no games provided")
			}

			ids := make([]types.ObjectId, 0, len(games))
			for _, g := range games {
				ids = append(ids, types.ObjectId(g))
			}

			data := struct {
				Games []types.ObjectId `json:"games"`
			}{
				Games: ids,
			}
			b, _ := json.Marshal(data)

			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/games/fen",
				httpAddress), bytes.NewBuffer(b))
			if err != nil {
				return err
			}

			return sendHttpRequest(httpAddress, req)
		}}

	cmd.Flags().Int64SliceVar(&games, "games", []int64{int64(types.NewObjectId())}, "Games IDs")

	return cmd
}
func getGamePGN(httpAddress string) *cobra.Command {
	var gameId int64

	cmd := &cobra.Command{
		Use:   "get_pgn",
		Short: "Send a GetGamePGN request",
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/games/pgn/%d",
				httpAddress, gameId), nil)
			if err != nil {
				return err
			}

			return sendHttpRequest(httpAddress, req)
		},
	}

	cmd.Flags().Int64Var(&gameId, "gameId", int64(types.NewObjectId()), "Game ID")

	return cmd
}

func sendHttpRequest(httpAddress string, r *http.Request) error {
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(b))

	return nil
}
