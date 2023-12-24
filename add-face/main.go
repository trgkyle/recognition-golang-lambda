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

	// Face kognition collection ID
	collectionID := "customers"

	input := &rekognition.IndexFacesInput{
		CollectionId: aws.String(collectionID),
		Image: &rekognition.Image{
			S3Object: &rekognition.S3Object{
				Bucket: aws.String(bucket),
				Name:   aws.String(key),
			},
		},
	}

	result, err := svc.IndexFaces(input)
	if err != nil {
		log.Printf("Error index faces :>> %v", err)
		return
	}

	userId := strings.TrimSuffix(key, filepath.Ext(key))

	associateFaces(svc, collectionID, userId, result.FaceRecords)

}

func associateFaces(svc *rekognition.Rekognition, collectionID string, userId string, faceRecords []*rekognition.FaceRecord) {
	faceIds := make([]*string, len(faceRecords))

	for i, face := range faceRecords {
		faceIds[i] = face.Face.FaceId
	}

	userCreated := &rekognition.CreateUserInput{
		CollectionId: aws.String(collectionID),
		UserId:       aws.String(userId),
	}

	svc.CreateUser(userCreated)

	input := &rekognition.AssociateFacesInput{
		CollectionId: aws.String(collectionID),
		UserId:       aws.String(userId),
		FaceIds:      faceIds,
	}

	_, err := svc.AssociateFaces(input)
	if err != nil {
		log.Printf("Error associating faces to user :>> %v", err)
		return
	}
}

func main() {
	lambda.Start(handler)
}
