package utils

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/Danny-Dasilva/CycleTLS/cycletls"
	"github.com/PuerkitoBio/goquery"
	"github.com/rodatboat/go-vocab/model"
)

func GetCookiesString(cookies []cycletls.Cookie) (string, error) {
	if cookies != nil {
		cookieHeader := ""
		for _, cookie := range cookies {
			cookieHeader += cookie.Name + "=" + cookie.Value + ";"
		}
		fmt.Println("Cookie header:", cookieHeader)
		return cookieHeader, nil
	}
	return "", errors.New("no cookies found")
}

func ExtractQuestion(data map[string]interface{}) (*model.Question, string, error) {
	question := model.Question{}
	secret, err := ExtractSecret(data)
	if err != nil {
		fmt.Println("Error extracting secret:", err)
		panic(err)
	}

	questionData, ok := data["question"].(map[string]interface{})
	if !ok {
		fmt.Println("Error getting question data, trying base data JSON instead...")
		questionData = data
		question.QuestionType = questionData["qtype"].(string)
	} else {
		question.QuestionType = questionData["type"].(string)
	}

	question.IsCorrect = false
	question.Code = questionData["code"].(string)
	question.Difficulty = questionData["difficulty"].(float64)

	decodedQuestion, err := base64.StdEncoding.DecodeString(question.Code)
	if err != nil {
		fmt.Println("Error base64 decoding question:", err)
		return nil, "", err
	}
	question.DecodedCode = string(decodedQuestion)

	// Create HTML doc using a string reader
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(question.DecodedCode))
	if err != nil {
		fmt.Println("Error creating HTML doc:", err)
		return nil, "", err
	}

	questionContext := doc.Find("div.questionContent").First()
	contextParts := questionContext.Find("div.sentence")
	questionContent := doc.Find("div.instructions").First()

	question.QuestionContext = stripExtraWhiteSpace(contextParts.Text())
	question.Question = stripExtraWhiteSpace(questionContent.Text())

	if question.QuestionType == "T" {
		spellingQuestionAnswer := doc.Find("div.complete strong").First().Text()
		question.Answer = spellingQuestionAnswer
		question.AnswerKey = spellingQuestionAnswer
	}

	var choices []model.QuestionChoices
	doc.Find("div.choices a").Each(func(i int, s *goquery.Selection) {
		keyVal, ok := s.Attr("data-nonce")
		if !ok {
			fmt.Println("Error getting data-nonce")
			return
		}
		val := stripExtraWhiteSpace(s.Text())
		if val == question.Answer {
			question.IsCorrect = true
			question.AnswerKey = keyVal
		}

		choices = append(choices, model.QuestionChoices{
			Key:   keyVal,
			Value: val,
		})
	})
	question.Choices = choices

	return &question, secret, nil
}

func stripExtraWhiteSpace(str string) string {
	trimmed := strings.TrimSpace(str)
	words := strings.Fields(trimmed)
	return strings.Join(words, " ")
}

func PrettyPrint(v interface{}) {
	// Marshal the struct to a JSON-formatted byte slice with indentation
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Convert the byte slice to a string and print it
	fmt.Println(string(jsonBytes))
}

func ExtractSecret(data map[string]interface{}) (string, error) {
	secret, ok := data["secret"].(string)
	if !ok {
		err := errors.New("error decoding secret")
		return "", err
	}

	// Store in progress, then load on startup
	return secret, nil
}

func ExtractPracticeProgress(data map[string]interface{}) (*float64, error) {
	gameJson, ok := data["game"].(map[string]interface{})
	if !ok {
		err := errors.New("failed to decode game JSON")
		return nil, err
	}
	progress, ok := gameJson["progress"].(float64)
	if !ok {
		err := errors.New("failed to decode progress JSON")
		return nil, err
	}
	return &progress, nil
}

func GenerateRandomTime() int {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	randomFloat := 3 + r.Float64()*(9-3)
	roundedFloat := math.Round(randomFloat*1000) / 1000
	result := int(roundedFloat * 1000)
	return result
}

func RetrieveCookies(cookies []*http.Cookie, existingCookies []cycletls.Cookie) []cycletls.Cookie {
	var newCookiesMap map[string]cycletls.Cookie = make(map[string]cycletls.Cookie)
	for _, existingCookie := range existingCookies {
		newCookiesMap[existingCookie.Name] = existingCookie
	}

	for _, cookie := range cookies {
		if isImportantCookie(cookie.Name) {
			newCookiesMap[cookie.Name] = cycletls.Cookie{
				Name:       cookie.Name,
				Value:      cookie.Value,
				Domain:     cookie.Domain,
				Path:       cookie.Path,
				Expires:    cookie.Expires,
				Secure:     cookie.Secure,
				RawExpires: cookie.RawExpires,
				MaxAge:     cookie.MaxAge,
				HTTPOnly:   cookie.HttpOnly,
				SameSite:   cookie.SameSite,
				Raw:        cookie.Raw,
				Unparsed:   cookie.Unparsed,
			}
		}
	}

	newCookies := make([]cycletls.Cookie, 0, len(cookies))
	for _, cookie := range newCookiesMap {
		newCookies = append(newCookies, cookie)
	}

	return newCookies
}

func isImportantCookie(cookie_name string) bool {
	IMPORTANT_COOKIES := []string{"AWSALB", "AWSALBCORS", "JSESSIONID", "guid", "__cf_bm"}
	for _, cookie := range IMPORTANT_COOKIES {
		if cookie == cookie_name {
			return true
		}
	}
	return false
}
