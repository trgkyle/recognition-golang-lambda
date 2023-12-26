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

		responseData, err := json.Marshal(result)
		if err != nil {
			log.Printf("Error marshaling data to JSON: %v", err)
			return err
		}

		log.Printf("Print response data %v", responseData)
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

func main() {
	lambda.Start(handler)
}
