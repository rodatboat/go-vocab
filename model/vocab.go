package model

type Question struct {
	QuestionType string
	DecodedCode  string
	Code         string
	Difficulty   float64

	QuestionContext string
	Question        string
	Answer          string
	AnswerKey       string
	Choices         []QuestionChoices

	IsCorrect  bool
	TargetWord string
}

type QuestionChoices struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type WordList struct {
	ID                   int
	Name                 string
	WordCount            int
	CompletionPercentage float64
	Completed            bool
}

type AnswerReq struct {
	Secret string `json:"secret"`
	V      int    `json:"v"`
	Rt     int    `json:"rt"`
	A      string `json:"a"`
}

type NextQuestionReq struct {
	Secret string `json:"secret"`
	V      int    `json:"v"`
}

type StartPracticeReq struct {
	V            int    `json:"v"`
	ActivityType string `json:"activitytype"`
	WordListId   int    `json:"wordlistid"`
	Secret       string `json:"secret,omitempty"`
}

type Cookies struct {
	AlbCookie  string `json:"AWSALB"`
	JSessionId string `json:"JSESSIONID"`
	Guid       string `json:"guid"`
}
