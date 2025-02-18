package utils

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
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
		return cookieHeader, nil
	}
	return "", errors.New("no cookies found")
}

func ExtractQuestion(data map[string]interface{}) (model.Question, string, error) {
	question := model.Question{}
	secret, ok := data["secret"].(string)
	if !ok {
		return model.Question{}, "", errors.New("secret not found")
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
		return model.Question{}, "", err
	}
	question.DecodedCode = string(decodedQuestion)

	// Create HTML doc using a string reader
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(question.DecodedCode))
	if err != nil {
		fmt.Println("Error creating HTML doc:", err)
		return model.Question{}, "", err
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

	return question, secret, nil
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
	return secret, nil
}

func ExtractPracticeProgress(data map[string]interface{}) (float64, error) {
	gameJson, ok := data["game"].(map[string]interface{})
	if !ok {
		err := errors.New("failed to decode game JSON")
		return 0, err
	}
	progress, ok := gameJson["progress"].(float64)
	if !ok {
		err := errors.New("failed to decode progress JSON")
		return 0, err
	}
	return progress, nil
}

func GenerateRandomTime() int {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	randomFloat := 3 + r.Float64()*(7-3)
	roundedFloat := math.Round(randomFloat*1000) / 1000
	result := int(roundedFloat * 1000)
	return result
}
