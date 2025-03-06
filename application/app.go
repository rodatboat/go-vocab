package application

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Danny-Dasilva/CycleTLS/cycletls"
	"github.com/jackc/pgx/v4"
	"github.com/rodatboat/go-vocab/model"
	"github.com/rodatboat/go-vocab/utils"
)

const CONTENT_TYPE_HEADER = "Content-Type"
const CONTENT_TYPE_URL_ENCODED = "application/x-www-form-urlencoded; charset=UTF-8"
const CONTENT_TYPE_JSON = "application/json; charset=UTF-8"

type RunParams struct {
	ListId int

	AlbCookie  string
	JSessionId string
	Guid       string

	Ja3 string
}

type RunContext struct {
	ListId                      int
	OllamaQuery                 string
	CurrentQuestion             *model.Question
	PointsEarned                int
	Secret                      string
	ErrorCount                  int
	CurrentCompletionPercentage float64

	Cookies []cycletls.Cookie
}

type RunDBConfig struct {
	dbname   string
	host     string
	port     string
	user     string
	password string
}

type Runner struct {
	DBConfig      RunDBConfig
	Conn          *pgx.Conn
	ctx           *RunContext
	client        cycletls.CycleTLS
	clientOptions cycletls.Options
}

func New(params RunParams) *Runner {
	cookies := []cycletls.Cookie{
		{Name: "AWSALB", Value: params.AlbCookie},
		{Name: "JSESSIONID", Value: params.JSessionId},
		{Name: "guid", Value: params.Guid},
	}

	cookieHeader, err := utils.GetCookiesString(cookies)
	if err != nil {
		fmt.Println("Error creating cookie header string", err)
		panic(err)
	}

	rawQuery, err := os.ReadFile("./db/ai_query.json")
	if err != nil {
		fmt.Println("Error reading ai_query.json:", err)
		panic(err)
	}

	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36 OPR/117.0.0.0"
	options := cycletls.Options{
		Ja3:       params.Ja3,
		UserAgent: userAgent,
		Headers: map[string]string{
			"User-Agent":       userAgent,
			"Origin":           "https://www.vocabulary.com",
			"X-Requested-With": "XMLHttpRequest",
			"Content-Type":     CONTENT_TYPE_URL_ENCODED,
			"Cookie":           cookieHeader,
		},
		Cookies: cookies,
	}

	runner := &Runner{
		DBConfig: RunDBConfig{
			dbname:   "vocabularycom",
			host:     "localhost",
			port:     "5432",
			user:     "postgres",
			password: "password",
		},
		ctx: &RunContext{
			ListId:      params.ListId,
			Cookies:     options.Cookies,
			OllamaQuery: string(rawQuery),
		},
		client:        cycletls.Init(),
		clientOptions: options,
	}
	runner.initDb(runner.DBConfig)

	return runner
}

func (r *Runner) IsLoggedIn() bool {
	res := false
	ME_URI := "https://www.vocabulary.com/auth/me.json"

	fmt.Println("Checking if logged in...")
	resp, err := r.client.Do(ME_URI, r.clientOptions, "GET")
	if err != nil {
		fmt.Println("Error making HTTP request:", err)
		res = false
	}

	data := resp.JSONBody()
	auth, ok := data["auth"].(map[string]interface{})
	if !ok {
		res = false
	}

	loggedIn, ok := auth["loggedin"].(bool)
	if ok {
		res = loggedIn
	}

	return res
}

func (r *Runner) Start(listId int) *model.Question {
	START_URI := "https://www.vocabulary.com/challenge/start.json"

	payload := model.StartPracticeReq{
		V:            3,
		ActivityType: "p",
		WordListId:   listId,
	}

	if r.ctx.Secret != "" {
		payload.Secret = r.ctx.Secret
		fmt.Println("Secret is not empty, continue from where we left off...")
	} else {
		r.ctx.CurrentCompletionPercentage = 0
	}

	// Set body
	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return nil
	}
	r.clientOptions.Body = string(body)
	r.clientOptions.Headers["Content-Type"] = CONTENT_TYPE_JSON

	fmt.Println("Starting practice session...")
	resp, err := r.client.Do(START_URI, r.clientOptions, "POST")
	if err != nil {
		fmt.Println("Error making HTTP request:", err)
		return nil
	}

	// Parse the JSON response
	data := resp.JSONBody()
	r.clientOptions.Cookies = utils.RetrieveCookies(resp.Cookies, r.clientOptions.Cookies)
	r.clientOptions.Headers["Cookie"], _ = utils.GetCookiesString(r.clientOptions.Cookies)
	if resp.Status == 400 {
		roundOver, ok := data["error"].(string)
		if ok {
			if roundOver == "RestartChallengeException" {
				fmt.Println("Encountered RestartChallengeException. Round over.")
				r.ctx.CurrentCompletionPercentage = 1
				r.ctx.CurrentQuestion = nil
				return nil
			}
		}
	}

	secret, err := utils.ExtractSecret(data)
	if err != nil {
		fmt.Println("Error extracting secret:", err)
		return nil
	}
	r.ctx.Secret = secret

	question, _, err := utils.ExtractQuestion(data)
	if err != nil {
		fmt.Println("Error extracting question:", err)
		return nil
	}
	r.ctx.CurrentQuestion = question
	r.SaveQuestionToDB(*question)

	progress, err := utils.ExtractPracticeProgress(data)
	if err != nil {
		fmt.Println("Error extracting progress, skipping...")
	} else {
		r.ctx.CurrentCompletionPercentage = *progress
	}

	return question
}

// Initializes db connection, and creates required tables.
func (r *Runner) initDb(config RunDBConfig) {
	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		config.user, config.password, config.host, config.port, config.dbname)

	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		panic(err)
	}
	r.Conn = conn

	query, err := os.ReadFile("./db/ddl.sql")
	if err != nil {
		fmt.Println("Error reading ddl.sql:", err)
		panic(err)
	}

	_, err = r.Conn.Exec(context.Background(), string(query))
	if err != nil {
		fmt.Println("Error executing ddl.sql:", err)
		panic(err)
	}
}

func (r *Runner) SaveQuestionToDB(question model.Question) {

	query := `
		INSERT INTO question (
			question_type,
			question,
			question_context,
			question_code,
			question_html,
			answer,
			answer_data_key,
			difficulty,
			choices,
			correct,
			target_word
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11    
		)
		ON CONFLICT (question_type, question_context, question) DO UPDATE SET
			answer = $6,
			answer_data_key = $7,
			correct = $10,
			target_word = $11
		WHERE question.correct = FALSE
	`

	choicesJson, err1 := json.Marshal(question.Choices)
	if err1 != nil {
		fmt.Println("Error marshaling JSON:", err1)
		choicesJson = nil
	}

	_, err := r.Conn.Exec(context.Background(), query,
		question.QuestionType,
		question.Question,
		question.QuestionContext,
		question.Code,
		question.DecodedCode,
		question.Answer,
		question.AnswerKey,
		question.Difficulty,
		choicesJson,
		question.IsCorrect,
		question.TargetWord)
	if err != nil {
		fmt.Println("Error executing question insert query:", err)
		panic(err)
	}
}

type OllamaPayload struct {
	Context  string                  `json:"context"`
	Question string                  `json:"question"`
	Choices  []model.QuestionChoices `json:"choices"`
}

func (r *Runner) Ask(question model.Question) model.QuestionChoices {
	LLM_URI := "http://localhost:11434/api/generate"

	payload := &OllamaPayload{
		Context:  question.QuestionContext,
		Question: question.Question,
		Choices:  question.Choices,
	}

	payloadJson, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		payloadJson = nil
	}
	payloadString := strings.ReplaceAll(string(payloadJson), "\"", "\\\"")
	query := fmt.Sprintf(r.ctx.OllamaQuery, payloadString)

	req, err := http.NewRequest("POST", LLM_URI, bytes.NewBuffer([]byte(query)))
	if err != nil {
		fmt.Println("Error creating request:", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request using an HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
	}

	// Parse the JSON response
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		fmt.Println("Error decoding JSON response:", err)
		panic(err)
	}

	response, ok := data["response"].(string)
	if !ok {
		err := errors.New("failed to decode response JSON")
		panic(err)
	}

	// Parse the JSON response
	var answerJson map[string]interface{}
	if err := json.Unmarshal([]byte(response), &answerJson); err != nil {
		fmt.Println("Error decoding JSON response:", err)
		panic(err)
	}

	innerAnswerJson, ok := answerJson["answer"].(map[string]interface{})
	if !ok {
		err := errors.New("failed to decode answer JSON")
		panic(err)
	}

	answer, ok := innerAnswerJson["answer"].(string)
	if !ok {
		err := errors.New("failed to decode inner answer JSON")
		panic(err)
	}

	code, ok := innerAnswerJson["code"].(string)
	if !ok {
		err := errors.New("failed to decode code JSON")
		panic(err)
	}

	r.ctx.CurrentQuestion.Answer = answer
	r.ctx.CurrentQuestion.AnswerKey = code
	return model.QuestionChoices{
		Key:   code,
		Value: answer,
	}
}

func (r *Runner) AnswerQuestion(answer model.QuestionChoices) {
	SAVE_ANSWER_URI := "https://www.vocabulary.com/challenge/saveanswer.json"
	// Send request, update secret, get next question after this method.
	requestPayload := model.AnswerReq{
		Secret: r.ctx.Secret,
		V:      3,
		Rt:     utils.GenerateRandomTime(),
		A:      answer.Key,
	}

	// Set body
	body, err := json.Marshal(requestPayload)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		panic(err)
	}
	r.clientOptions.Body = string(body)
	r.clientOptions.Headers["Content-Type"] = CONTENT_TYPE_URL_ENCODED

	fmt.Println("Answering question...")
	resp, err := r.client.Do(SAVE_ANSWER_URI, r.clientOptions, "POST")
	if err != nil {
		fmt.Println("Error making HTTP request:", err)
		panic(err)
	}

	// Parse the JSON response
	data := resp.JSONBody()
	r.clientOptions.Cookies = utils.RetrieveCookies(resp.Cookies, r.clientOptions.Cookies)
	r.clientOptions.Headers["Cookie"], _ = utils.GetCookiesString(r.clientOptions.Cookies)
	if resp.Status == 400 {
		roundOver, ok := data["error"].(string)
		if ok {
			if roundOver == "RestartChallengeException" {
				fmt.Println("Encountered RestartChallengeException. Round over.")
				r.ctx.CurrentCompletionPercentage = 1
				r.ctx.CurrentQuestion = &model.Question{}
				return
			}
		}
	}

	secret, err := utils.ExtractSecret(data)
	if err != nil {
		fmt.Println("Error extracting secret:", err)
		panic(err)
	}
	r.ctx.Secret = secret

	answerJson, ok := data["answer"].(map[string]interface{})
	if !ok {
		err := errors.New("failed to decode answer JSON")
		panic(err)
	}
	wasCorrect, ok := answerJson["correct"].(bool)
	if !ok {
		err := errors.New("failed to decode wasCorrect JSON")
		panic(err)
	}
	targetWord, ok := answerJson["word"].(string)
	if !ok {
		err := errors.New("failed to decode target word JSON")
		panic(err)
	}

	r.ctx.Secret = secret

	r.ctx.CurrentQuestion.Answer = answer.Value
	r.ctx.CurrentQuestion.AnswerKey = answer.Key
	r.ctx.CurrentQuestion.TargetWord = targetWord
	r.ctx.CurrentQuestion.IsCorrect = wasCorrect
	r.ctx.PointsEarned = answerJson["points"].(int) + answerJson["bonus"].(int)

	r.SaveQuestionToDB(*r.ctx.CurrentQuestion)

	progress, err := utils.ExtractPracticeProgress(data)
	if err != nil {
		fmt.Println("Error extracting progress:", err)
		panic(err)
	}
	r.ctx.CurrentCompletionPercentage = *progress
}

func (r *Runner) NextQuestion() *model.Question {
	// To be called after answerQuestion()
	NEXT_QUESTION_URI := "https://www.vocabulary.com/challenge/nextquestion.json"
	requestPayload := model.NextQuestionReq{
		Secret: r.ctx.Secret,
		V:      3,
	}

	// Set body
	body, err := json.Marshal(requestPayload)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		panic(err)
	}
	r.clientOptions.Body = string(body)
	r.clientOptions.Headers["Content-Type"] = CONTENT_TYPE_URL_ENCODED

	fmt.Println("Fetching next question...")
	resp, err := r.client.Do(NEXT_QUESTION_URI, r.clientOptions, "POST")
	if err != nil {
		fmt.Println("Error making HTTP request:", err)
		panic(err)
	}

	// Parse the JSON response
	data := resp.JSONBody()
	r.clientOptions.Cookies = utils.RetrieveCookies(resp.Cookies, r.clientOptions.Cookies)
	r.clientOptions.Headers["Cookie"], _ = utils.GetCookiesString(r.clientOptions.Cookies)
	secret, err := utils.ExtractSecret(data)
	if err != nil {
		fmt.Println("Error extracting secret:", err)
		panic(err)
	} else {
		r.ctx.Secret = secret
	}

	question, _, _ := utils.ExtractQuestion(data)
	r.ctx.CurrentQuestion = question
	r.SaveQuestionToDB(*question)

	progress, err := utils.ExtractPracticeProgress(data)
	if err != nil {
		fmt.Println("Error extracting progress, skipping...")
	} else {
		r.ctx.CurrentCompletionPercentage = *progress
	}

	return question
}

func (r *Runner) Practice() {
	// r.ctx.CurrentQuestion = r.Start(r.ctx.ListId)
	r.ctx.CurrentQuestion = r.NextQuestion()
	for {
		answer := r.Ask(*r.ctx.CurrentQuestion)

		time.Sleep(3 * time.Second)
		r.AnswerQuestion(answer)

		if r.ctx.CurrentCompletionPercentage == 1 {
			fmt.Println("Round over. Restarting challenge...")
			r.ctx.Secret = ""
			r.ctx.CurrentQuestion = r.Start(r.ctx.ListId)

			time.Sleep(3 * time.Second)
			r.AnswerQuestion(answer)
		}

		time.Sleep(3 * time.Second)
		r.ctx.CurrentQuestion = r.NextQuestion()

		fmt.Println("Sleeping for 3 seconds...")
		time.Sleep(3 * time.Second)
	}
}
