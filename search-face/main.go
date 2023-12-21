package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/rekognition"
)

func handler(ctx context.Context, s3Event events.S3Event) {
	sess := session.Must(session.NewSession())

	// Iterate over each S3 event
	for _, record := range s3Event.Records {
		s3 := record.S3
		key := s3.Object.Key
		bucket := s3.Bucket.Name

		// Perform facial recognition using AWS Rekognition
		searchFaces(sess, bucket, key)
	}
}

func searchFaces(sess *session.Session, bucket, key string) {
	svc := rekognition.New(sess)

	input := &rekognition.SearchFacesByImageInput{
		CollectionId: aws.String("customers"),
		Image: &rekognition.Image{
			S3Object: &rekognition.S3Object{
				Bucket: aws.String(bucket),
				Name:   aws.String(key),
			},
		},
	}

	result, err := svc.SearchFacesByImage(input)
	if err != nil {
		log.Printf("Error index faces:>> %v", err)
		return
	}

	searchFacesInDynamoDB(sess, result.FaceMatches)
}

func searchFacesInDynamoDB(sess *session.Session, faceDetails []*rekognition.FaceMatch) {
	dbSvc := dynamodb.New(sess)

	for _, match := range faceDetails {
		faceId := *match.Face.FaceId

		// Retrieve identity information from DynamoDB
		getItemInput := &dynamodb.GetItemInput{
			TableName: aws.String("face-recognition-authenticated"), // Replace with your DynamoDB table name
			Key: map[string]*dynamodb.AttributeValue{
				"faceId": {
					S: aws.String(faceId),
				},
			},
		}

		getItemOutput, err := dbSvc.GetItem(getItemInput)
		if err != nil {
			log.Println("Error retrieving item from DynamoDB:", err.Error())
			continue // Skip to the next faceMatch in case of errors
		}

		if getItemOutput.Item != nil {
			// Identity information found
			name := *getItemOutput.Item["customerId"].S

			log.Printf("Found customer ID:>> %v", name)
		} else {
			log.Printf("Warning: Face ID not found in DynamoDB :>> %v", faceId)
		}
	}

}

func main() {
	lambda.Start(handler)
}
