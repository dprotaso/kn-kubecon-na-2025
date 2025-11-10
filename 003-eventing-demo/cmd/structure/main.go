package main

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"kndemo"
	"log"
	"os"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func GenerateSchema[T any]() any {
	// Structured Outputs uses a subset of JSON schema
	// These flags are necessary to comply with the subset
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

var ParsedDataSchema = GenerateSchema[kndemo.ParsedData]()

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
	var payload *kndemo.Payload

	if err := e.DataAs(&payload); err != nil {
		return nil, fmt.Errorf("unable to unmarshal data: %w", err)
	}

	if payload.Content == "" {
		return nil, errors.New("payload missing content")
	}

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "parse_data",
		Description: openai.String("Notable information parsed about the response"),
		Schema:      ParsedDataSchema,
		Strict:      openai.Bool(true),
	}

	chat, err := r.client.Chat.Completions.New(ctx,
		openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage(
					` You are a world-class text analysis expert. 
					  Extract the information precisely into the provided JSON format. 
						The message is a customer support email.
					`,
				),
				openai.UserMessage(payload.Content),
			},
			Model: envOr("LLM_MODEL", "granite3.3:8b"),
			ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
				OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{JSONSchema: schemaParam},
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
	event.SetSource("structure")
	event.SetType("structure.new")

	var data kndemo.ParsedData
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), &data)

	if err != nil {
		log.Println("failed to parse LLM json", err)
		event.SetType("structure.failure")
		event.SetData(cloudevents.ApplicationJSON, map[string]any{"error": err.Error()})
	} else {
		payload.ParsedData = &data
	}

	event.SetData(cloudevents.ApplicationJSON, payload)

	return &event, nil
}

func envOr(key, defaultVal string) string {
	return cmp.Or(os.Getenv(key), defaultVal)
}
