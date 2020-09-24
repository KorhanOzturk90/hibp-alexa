package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/ericdaugherty/alexa-skills-kit-golang"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

var (
	alexaMetaData = &alexa.Alexa{ApplicationID: "amzn1.ask.skill.fc9cb575-acd5-4d3c-9145-669c62cc4363",
		RequestHandler: &requestHandler{}, IgnoreApplicationID: true, IgnoreTimestamp: true}
)

type requestHandler struct{}

type breach struct {
	Name string
}

func main() {
	lambda.Start(Handle)
}

func Handle(ctx context.Context, requestEnv *alexa.RequestEnvelope) (interface{}, error) {
	return alexaMetaData.ProcessRequest(ctx, requestEnv)
}

func (r requestHandler) OnSessionStarted(context context.Context, request *alexa.Request, session *alexa.Session, aContext *alexa.Context, response *alexa.Response) error {
	fmt.Printf("OnSessionStarted requestId=%s, sessionId=%s", request.RequestID, session.SessionID)

	return nil
}

func (r requestHandler) OnLaunch(context context.Context, request *alexa.Request, session *alexa.Session, aContext *alexa.Context, response *alexa.Response) error {
	speechText := "Welcome to have I been hacked. "
	fmt.Printf("OnLaunch requestId=%s, sessionId=%s", request.RequestID, session.SessionID)

	response.SetSimpleCard("Am I hacked", speechText)
	response.SetOutputText(speechText)
	response.SetRepromptText(speechText)

	response.ShouldSessionEnd = false
	return nil
}

func (r requestHandler) OnIntent(context context.Context, request *alexa.Request, session *alexa.Session, aContext *alexa.Context, response *alexa.Response) error {
	fmt.Printf("OnIntent requestId=%s, sessionId=%s, intent=%s", request.RequestID, session.SessionID, request.Intent.Name)

	switch request.Intent.Name {
	case "breachIntent":
		fmt.Println("AmIHackedIntent triggered")

		endpoint := aContext.System.APIEndpoint
		apiAccessToken := aContext.System.APIAccessToken

		breachesFound := isEmailCompromised(getUserEmail(endpoint, apiAccessToken))

		var speechText strings.Builder
		if len(breachesFound) == 0 {
			speechText.WriteString("Your email has not been involved in any data breaches. Yay!")
		} else {
			speechText.WriteString("Your email was included in the following breaches: ")
			for _, b := range breachesFound {
				fmt.Printf("email found in %v breach", b)
				speechText.WriteString(b.Name + ", ")
			}
		}

		response.SetSimpleCard("Am I hacked", speechText.String())
		response.SetOutputText(speechText.String())
		response.ShouldSessionEnd = true

	case "AMAZON.HelpIntent":
		fmt.Println("AMAZON.HelpIntent triggered")
		speechText := "You can say hello to me!"

		response.SetSimpleCard("HelloWorld", speechText)
		response.SetOutputText(speechText)
		response.SetRepromptText(speechText)
	default:
		return errors.New("Invalid Intent")
	}

	return nil
}

func getUserEmail(endpoint string, apiAccessToken string) string {
	client := &http.Client{}
	url := endpoint + "/v2/accounts/~current/settings/Profile.email"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Print(err)
	}
	req.Header.Add("Authorization", "Bearer "+apiAccessToken)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("user email: %v", string(body))
	return trimEmail(body)
}

func isEmailCompromised(email string) []breach {
	client := &http.Client{}
	url := "https://haveibeenpwned.com/api/v3/breachedaccount/" + email
	var breachList []breach
	req, err := http.NewRequest("GET", url, nil)

	req.Header.Add("hibp-api-key", os.Getenv("PWNED_API_KEY"))
	fmt.Printf("hibp request %v", req)
	resp, err := client.Do(req)

	if err != nil {
		fmt.Print(err)
	}
	defer resp.Body.Close()
	fmt.Printf("hibp response %v", resp)
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("body: %v", string(body))
	jsonErr := json.Unmarshal(body, &breachList)
	if jsonErr != nil {
		fmt.Print(err)
	}
	fmt.Printf("breaches found: %v", breachList)

	return breachList
}

func trimEmail(email []byte) string {
	trimmedEmail := email[1 : len(email)-1]
	fmt.Printf("trimmed email: %v", string(trimmedEmail))
	return string(trimmedEmail)
}

func (r *requestHandler) OnSessionEnded(context context.Context, request *alexa.Request, session *alexa.Session, aContext *alexa.Context, response *alexa.Response) error {

	fmt.Printf("OnSessionEnded requestId=%s, sessionId=%s", request.RequestID, session.SessionID)

	return nil
}
