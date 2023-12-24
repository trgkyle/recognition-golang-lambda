package main

import (
	"context"
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

		// Search for users using the image
		searchUsersByImage(sess, bucket, key)
	}
}

func searchUsersByImage(sess *session.Session, bucket, key string) {
	svc := rekognition.New(sess)

	collectionID := "customers"
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
		log.Printf("Error searching for users by image :>> %v", err)
		return
	}

	log.Printf("Found %d user matches 🧐:", len(result.UserMatches))
	for _, match := range result.UserMatches {
		userID := *match.User.UserId
		similarity := *match.Similarity
		log.Printf("- User ID: %s (Similarity: %.2f%%)", userID, similarity)
	}
}

func main() {
	lambda.Start(handler)
}
