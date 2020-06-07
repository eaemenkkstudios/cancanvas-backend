package service

import (
	"bytes"
	"log"
	"os"

	"github.com/99designs/gqlgen/graphql"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AwsService interface
type AwsService interface {
	UploadFile(upload graphql.Upload, prefix string) (string, error)
}

type awsService struct {
	session    *session.Session
	awsRegion  string
	secretID   string
	secretKey  string
	bucketName string
	urlPrefix  string
}

func (a *awsService) UploadFile(upload graphql.Upload, prefix string) (string, error) {
	if prefix != "" {
		prefix = prefix + "/"
	}
	buffer := make([]byte, upload.Size)
	upload.File.Read(buffer)
	var tempFileName string = prefix + primitive.NewObjectID().Hex() + "-" + upload.Filename
	_, err := s3.New(a.session).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(a.bucketName),
		Key:                  aws.String(tempFileName),
		ACL:                  aws.String("public-read"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(upload.Size),
		ContentType:          aws.String(upload.ContentType),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})
	if err != nil {
		return "", err
	}
	return a.urlPrefix + tempFileName, nil
}

// NewAwsService function
func NewAwsService() AwsService {
	awsRegion := os.Getenv("AWS_REGION")
	secretID := os.Getenv("AWS_SECRET_ID")
	secretKey := os.Getenv("AWS_SECRET_KEY")
	bucketName := os.Getenv("AWS_BUCKET_NAME")
	var urlPrefix string = "https://" + bucketName + ".s3." + awsRegion + ".amazonaws.com/"
	session, err := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
		Credentials: credentials.NewStaticCredentials(
			secretID,
			secretKey,
			"",
		),
	})
	if err != nil {
		log.Fatal(err)
	}
	return &awsService{
		session,
		awsRegion,
		secretID,
		secretKey,
		bucketName,
		urlPrefix,
	}
}
