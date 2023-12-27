package main

import (
	"context"
	"encoding/json"
	"log"
	"path/filepath"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/rekognition"
)

const collectionID = "customers"

type RecognitionResult struct {
	UniqueID    string                   `json:"uniqueID"`
	UserMatches []map[string]interface{} `json:"userMatches"`
}

func handler(ctx context.Context, s3Event events.S3Event) error {
	sess := session.Must(session.NewSession())

	for _, record := range s3Event.Records {
		s3 := record.S3
		key := s3.Object.Key
		bucket := s3.Bucket.Name
		uniqueID := strings.TrimSuffix(key, filepath.Ext(key))

		userMatches, err := searchUsersByImage(sess, bucket, key)
		if err != nil {
			log.Printf("Error searching for users by image: %v", err)
			continue
		}

		var userMatchesExtract []map[string]interface{}

		for _, match := range userMatches {
			userID := *match.User.UserId
			similarity := *match.Similarity

			// Create a map with user information
			userInfo := map[string]interface{}{
				"userID":     userID,
				"similarity": similarity,
			}
			userMatchesExtract = append(userMatchesExtract, userInfo)
		}

		result := RecognitionResult{
			UniqueID:    uniqueID,
			UserMatches: userMatchesExtract,
		}

		sendMessageToConnection(sess, result)
	}

	return nil
}

func searchUsersByImage(sess *session.Session, bucket, key string) ([]*rekognition.UserMatch, error) {
	svc := rekognition.New(sess)

	input := &rekognition.SearchUsersByImageInput{
		CollectionId: aws.String(collectionID),
		Image: &rekognition.Image{
			S3Object: &rekognition.S3Object{
				Bucket: aws.String(bucket),
				Name:   aws.String(key),
			},
		},
		MaxUsers: aws.Int64(5),
	}

	result, err := svc.SearchUsersByImage(input)
	if err != nil {
		return nil, err
	}

	return result.UserMatches, nil
}

type ConnectionItem struct {
	ConnectionID string `json:"connectionId"`
}

func sendMessageToConnection(sess *session.Session, recognitionResult RecognitionResult) error {
	endpoint := "https://k5nbcqn768.execute-api.ap-southeast-1.amazonaws.com/v1/"
	apigatewaymanagementapiSvc := apigatewaymanagementapi.New(sess, aws.NewConfig().WithEndpoint(endpoint))
	recognitionResultJSON, _ := json.Marshal(recognitionResult)

	// Create a DynamoDB session
	dynamoSess := session.Must(session.NewSession())
	dynamoSvc := dynamodb.New(dynamoSess)

	// Define the input for the Scan operation on the DynamoDB table
	scanInput := &dynamodb.ScanInput{
		TableName: aws.String("WebsocketAPIConnections"),
	}

	// Perform the Scan operation
	scanOutput, err := dynamoSvc.Scan(scanInput)
	if err != nil {
		log.Printf("Error scanning DynamoDB: %v", err)
		return err
	}

	// Iterate over each item (connection) in the DynamoDB table
	for _, item := range scanOutput.Items {
		connectionItem := ConnectionItem{}
		err = dynamodbattribute.UnmarshalMap(item, &connectionItem)
		if err != nil {
			log.Printf("Error unmarshalling item from DynamoDB: %v", err)
			continue
		}

		// Send the message to the connection
		input := &apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: aws.String(connectionItem.ConnectionID),
			Data:         recognitionResultJSON,
		}

		_, err = apigatewaymanagementapiSvc.PostToConnection(input)
		if err != nil {
			log.Printf("Error posting to connection: %v", err)
			continue
		}
	}

	return nil
}
func main() {
	lambda.Start(handler)
}
