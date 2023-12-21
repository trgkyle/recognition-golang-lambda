package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
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
		detectFaces(sess, bucket, key)
	}
}

func detectFaces(sess *session.Session, bucket, key string) {
	svc := rekognition.New(sess)

	input := &rekognition.DetectFacesInput{
		Image: &rekognition.Image{
			S3Object: &rekognition.S3Object{
				Bucket: aws.String(bucket),
				Name:   aws.String(key),
			},
		},
	}

	result, err := svc.DetectFaces(input)
	if err != nil {
		log.Printf("Error detecting faces: %v", err)
		return
	}

	// Process the result (e.g., print facial details)
	for _, faceDetail := range result.FaceDetails {
		fmt.Printf("Detected face: %+v\n", faceDetail)
	}
}

func main() {
	lambda.Start(handler)
}
