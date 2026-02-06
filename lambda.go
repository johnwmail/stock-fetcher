//go:build lambda

package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

var httpAdapter *httpadapter.HandlerAdapter

func init() {
	log.Println("Lambda cold start")
	server := NewServer("0")
	httpAdapter = httpadapter.New(server.Handler())
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return httpAdapter.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}
