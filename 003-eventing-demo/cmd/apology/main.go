package main

import (
	"cmp"
	"context"
	"fmt"
	"kndemo"
	"log"
	"os"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type Receiver struct {
	client openai.Client
}

func main() {
	r := &Receiver{
		client: openai.NewClient(
			option.WithBaseURL(envOr("LLM_API_BASE_URL", "http://localhost:11434/v1")),
			option.WithAPIKey(envOr("LLM_API_KEY", "ollama")),
		),
	}

	// The default client is HTTP.
	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}
	if err = c.StartReceiver(context.Background(), r.Receive); err != nil {
		log.Fatalf("failed to start receiver: %v", err)
	}
}

func (r *Receiver) Receive(ctx context.Context, e cloudevents.Event) (*cloudevents.Event, error) {
	fmt.Println("got event")
	var payload *kndemo.Payload

	if err := e.DataAs(&payload); err != nil {
		return nil, fmt.Errorf("unable to unmarshal data: %w", err)
	}

	fmt.Println("invoking llm")
	chat, err := r.client.Chat.Completions.New(ctx,
		openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(
					` You are a world-class sarcasitic and rude apologizer. 
					  Read the data and the information about the customer and apologize.
					`,
				),
				openai.UserMessage(string(e.Data())),
			},
			Model: envOr("LLM_MODEL", "granite3.3:8b"),
			ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
				OfText: &openai.ResponseFormatTextParam{Type: "text"},
			},
		})
	if err != nil {
		log.Println("failed to do LLM magic ", err)
		return nil, err
	}

	event := cloudevents.NewEvent()
	event.SetID(uuid.New().String())

	// This ID serves as the unique identifier for the message
	event.SetSubject(e.Subject())
	event.SetSource("apology")
	event.SetType("apology.new")

	payload.Apology = chat.Choices[0].Message.Content
	fmt.Println("got apology")

	if err != nil {
		log.Println("failed to parse LLM json", err)
		event.SetType("structure.failure")
		event.SetData(cloudevents.ApplicationJSON, map[string]any{"error": err.Error()})
	} else {
		event.SetData(cloudevents.ApplicationJSON, payload)
	}
	return &event, nil
}

func envOr(key, defaultVal string) string {
	return cmp.Or(os.Getenv(key), defaultVal)
}
