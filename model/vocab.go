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
	Secret string `form:"secret" json:"secret,omitempty"`
	V      int    `form:"v" json:"v"`
	Rt     int    `form:"rt" json:"rt"`
	A      string `form:"a" json:"a"`
}

type NextQuestionReq struct {
	Secret string `form:"secret" json:"secret,omitempty"`
	V      int    `form:"v" json:"v"`
}

type StartPracticeReq struct {
	V            int    `form:"v" json:"v"`
	ActivityType string `form:"activitytype" json:"activitytype"`
	WordListId   int    `form:"wordlistid" json:"wordlistid"`
	Secret       string `form:"secret" json:"secret,omitempty"`
}

type Cookies struct {
	AlbCookie  string `json:"AWSALB"`
	JSessionId string `json:"JSESSIONID"`
	Guid       string `json:"guid"`
}
