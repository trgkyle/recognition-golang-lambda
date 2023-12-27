package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type ConnectionItem struct {
	ConnectionID string `json:"connectionId"`
}

func handleRequest(event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	switch event.RequestContext.EventType {
	case "CONNECT":
		// Handle connect event
		item := ConnectionItem{
			ConnectionID: event.RequestContext.ConnectionID,
		}

		av, err := dynamodbattribute.MarshalMap(item)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500}, err
		}

		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String("WebsocketAPIConnections"),
		}

		_, err = svc.PutItem(input)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500}, err
		}

	case "DISCONNECT":
		// Handle disconnect event
		input := &dynamodb.DeleteItemInput{
			Key: map[string]*dynamodb.AttributeValue{
				"connectionId": {
					S: aws.String(event.RequestContext.ConnectionID),
				},
			},
			TableName: aws.String("WebsocketAPIConnections"),
		}

		_, err := svc.DeleteItem(input)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500}, err
		}
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func main() {
	lambda.Start(handleRequest)
}
