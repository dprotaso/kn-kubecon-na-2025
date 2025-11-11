package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"knative.dev/pkg/environment"
	"knative.dev/pkg/kmeta"

	serving "knative.dev/serving/pkg/client/clientset/versioned"
	"knative.dev/serving/pkg/client/informers/externalversions"
)

const (
	inputSchemaAnnotation  = "serving.knative.dev/mcp-tool-input"
	outputSchemaAnnotation = "serving.knative.dev/mcp-tool-output"
	titleAnnotation        = "serving.knative.dev/mcp-tool-title"
)

type Clients struct {
	K8s     kubernetes.Interface
	Serving serving.Interface
}

func createRestConfig() (*rest.Config, error) {
	env := new(environment.ClientConfig)
	env.InitFlags(flag.CommandLine)
	klog.InitFlags(flag.CommandLine)
	flag.Parse()
	return env.GetRESTConfig()
}

func createClients() Clients {
	cfg, err := createRestConfig()
	if err != nil {
		panic(err)
	}

	return Clients{
		K8s:     kubernetes.NewForConfigOrDie(cfg),
		Serving: serving.NewForConfigOrDie(cfg),
	}
}

func startInformers(ctx context.Context, server *mcp.Server) {
	c := createClients()

	sf := externalversions.NewSharedInformerFactoryWithOptions(
		c.Serving,
		10*time.Hour,
		externalversions.WithNamespace(namespace),
	)

	_, err := sf.Serving().V1().Services().Informer().AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
		AddFunc: func(obj any, isInInitialList bool) {
			addTool(server, obj)
		},
		UpdateFunc: func(oldObj, newObj any) {
			addTool(server, newObj)
		},
		DeleteFunc: func(obj any) {
			removeTool(server, obj)
		},
	})
	if err != nil {
		panic(err)
	}

	sf.Start(ctx.Done())
	sf.WaitForCacheSync(ctx.Done())
}

func removeTool(server *mcp.Server, obj any) {
	o, err := kmeta.DeletionHandlingAccessor(obj)
	if err != nil {
		fmt.Println("error getting accessor", err)
		return
	}
	fmt.Println("removing tool", o.GetName())
	server.RemoveTools(o.GetName())
}

func addTool(server *mcp.Server, obj any) {
	var (
		input, output jsonschema.Schema
		title         string
	)

	o, err := kmeta.DeletionHandlingAccessor(obj)
	if err != nil {
		fmt.Println("error getting accessor", err)
		return
	}

	if val, ok := o.GetAnnotations()[inputSchemaAnnotation]; !ok {
		// No schema skip adding tool as input schema is expected
		return
	} else if err := input.UnmarshalJSON([]byte(val)); err != nil {
		fmt.Println("error parsing input schema", err)
		return
	}

	if val, ok := o.GetAnnotations()[outputSchemaAnnotation]; !ok {
		// No schema skip adding as tool
		return
	} else if err := output.UnmarshalJSON([]byte(val)); err != nil {
		fmt.Println("error parsing output schema", err)
		return
	}

	if val, ok := o.GetAnnotations()[titleAnnotation]; ok {
		title = val
	}

	tool := &mcp.Tool{
		Name:         o.GetName(),
		Title:        title,
		InputSchema:  &input,
		OutputSchema: &output,
	}

	fmt.Println("adding/updating tool", o.GetName())
	server.AddTool(tool, KnativeServiceHandler)
}

func KnativeServiceHandler(ctx context.Context, r *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	url := fmt.Sprintf("http://%s.%s.svc", r.Params.Name, namespace)

	resp, err := http.Post(url, "application/json", bytes.NewReader(r.Params.Arguments))
	if err != nil {
		return nil, fmt.Errorf("failed to post body: %w", err)
	}

	defer resp.Body.Close() //nolint

	bytes, err := io.ReadAll(resp.Body)
	fmt.Println("tool result", string(bytes))

	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	return &mcp.CallToolResult{StructuredContent: json.RawMessage(bytes)}, nil
}
