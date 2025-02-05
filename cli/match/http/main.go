package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/alikarimi999/shahboard/types"
	"github.com/spf13/cobra"
)

func main() {
	var httpAddress string

	var rootCmd = &cobra.Command{
		Use:   "cli-tool",
		Short: "CLI tool for sending requests to HTTP server",
	}

	rootCmd.PersistentFlags().StringVar(&httpAddress, "http", "localhost:8082", "HTTP server address")

	rootCmd.AddCommand(sendMatchRequest(httpAddress))

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
	}
}

func sendMatchRequest(httpAddress string) *cobra.Command {
	var userId, scale int64

	cmd := &cobra.Command{
		Use:   "match",
		Short: "Send a new match request",
		RunE: func(cmd *cobra.Command, args []string) error {
			if scale == 0 {
				req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/user/match", httpAddress), nil)

				b64, _ := json.Marshal(types.User{ID: types.ObjectId(userId)})

				encoded := base64.StdEncoding.EncodeToString(b64)
				req.Header.Set("X-User-Data", encoded)

				return sendHttpRequest(req)
			}

			sendScaleMatchRequests(httpAddress, scale)
			return nil
		},
	}

	cmd.Flags().Int64Var(&userId, "user", types.NewObjectId().Int64(), "User ID")
	cmd.Flags().Int64Var(&scale, "scale", 0, "Scale sends multiple match requests")

	return cmd
}

func sendScaleMatchRequests(httpAddress string, scale int64) {
	wg := sync.WaitGroup{}

	for i := 0; i < int(scale); i++ {
		if i%100 == 0 {
			time.Sleep(1 * time.Second)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/user/match", httpAddress), nil)
			b64, _ := json.Marshal(types.User{ID: types.NewObjectId()})
			encoded := base64.StdEncoding.EncodeToString(b64)
			req.Header.Set("X-User-Data", encoded)
			if err := sendHttpRequest(req); err != nil {
				fmt.Println(err)
			}
		}()
	}
	wg.Wait()
	fmt.Println("Done!")
}

func sendHttpRequest(r *http.Request) error {
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
