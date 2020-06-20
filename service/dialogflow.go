package service

import (
	"context"
	"fmt"
	"log"
	"os"

	dialogflow "cloud.google.com/go/dialogflow/apiv2"
	"google.golang.org/api/option"
	dialogflowpb "google.golang.org/genproto/googleapis/cloud/dialogflow/v2"
)

// DialogflowService interface
type DialogflowService interface {
	SendMessage(sender, msg string) (string, error)
}

type dialogflowService struct {
	sessionsClient *dialogflow.SessionsClient
	projectID      string
}

func (s *dialogflowService) SendMessage(sender, msg string) (string, error) {
	sessionPath := fmt.Sprintf("projects/%s/agent/sessions/%s", s.projectID, sender)
	textInput := dialogflowpb.TextInput{Text: msg, LanguageCode: "en"}
	query := dialogflowpb.QueryInput_Text{Text: &textInput}
	queryInput := dialogflowpb.QueryInput{Input: &query}
	request := dialogflowpb.DetectIntentRequest{
		Session:    sessionPath,
		QueryInput: &queryInput,
	}
	response, err := s.sessionsClient.DetectIntent(context.TODO(), &request)
	if err != nil {
		return "", err
	}
	queryResult := response.GetQueryResult()
	fulfillmentText := queryResult.GetFulfillmentText()
	return fulfillmentText, nil
}

// NewDialogflowService function
func NewDialogflowService() DialogflowService {
	projectID := os.Getenv("DIALOGFLOW_PROJECT_ID")
	serviceAccount := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	sessionsClient, err := dialogflow.NewSessionsClient(
		context.TODO(),
		option.WithCredentialsJSON([]byte(serviceAccount)),
	)
	if err != nil {
		log.Fatal(err)
	}
	return &dialogflowService{
		sessionsClient,
		projectID,
	}
}
