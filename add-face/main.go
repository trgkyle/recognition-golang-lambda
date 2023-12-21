package main

import (
	"context"
	"log"
	"path/filepath"
	"strings"

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
		addFaces(sess, bucket, key)
	}
}

func addFaces(sess *session.Session, bucket, key string) {
	svc := rekognition.New(sess)

	input := &rekognition.IndexFacesInput{
		CollectionId: aws.String("customers"),
		Image: &rekognition.Image{
			S3Object: &rekognition.S3Object{
				Bucket: aws.String(bucket),
				Name:   aws.String(key),
			},
		},
	}

	result, err := svc.IndexFaces(input)
	if err != nil {
		log.Printf("Error index faces:>> %v", err)
		return
	}

	addFacesToDynamoDB(sess, result.FaceRecords, key)
}

func addFacesToDynamoDB(sess *session.Session, faceDetails []*rekognition.FaceRecord, key string) {
	dbSvc := dynamodb.New(sess)

	fileName := strings.TrimSuffix(key, filepath.Ext(key))

	// Modify this part based on your DynamoDB schema
	for _, faceDetail := range faceDetails {
		input := &dynamodb.PutItemInput{
			Item: map[string]*dynamodb.AttributeValue{
				"faceId": {
					S: aws.String(*faceDetail.Face.FaceId),
				},
				"customerId": {
					S: aws.String(fileName),
				},
			},
			TableName: aws.String("face-recognition-authenticated"),
		}

		_, err := dbSvc.PutItem(input)
		if err != nil {
			log.Printf("Error put to DynamoDB :>> %v", err)
			return
		}
	}
}

func main() {
	lambda.Start(handler)
}
