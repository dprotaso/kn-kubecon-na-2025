package main

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var namespace = "default"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	namespace = cmp.Or(os.Getenv("POD_NAMESPACE"), namespace)

	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "kubecon",
			Version: "v1.0.0",
		},
		&mcp.ServerOptions{
			GetSessionID: func() string {
				return "" // disable sessions
			},
		},
	)

	startInformers(ctx, mcpServer)

	handler := mcp.NewStreamableHTTPHandler(
		func(req *http.Request) *mcp.Server { return mcpServer },
		&mcp.StreamableHTTPOptions{
			Stateless: true,
		},
	)

	handlerWithLogging := loggingHandler(corsHandler(handler))

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", cmp.Or(os.Getenv("PORT"), "8080")),
		Handler: handlerWithLogging,
	}

	go func() {
		fmt.Printf("MCP server listening on %s\n", httpServer.Addr)

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Errorf("server failed: %w", err))
		}
	}()

	<-ctx.Done()

	sctx, scancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer scancel()
	if err := httpServer.Shutdown(sctx); err != nil {
		fmt.Println("Failed to shutdown server", err)
	}
}
