package model

type Prompt struct {
	Context  string            `json:"context"`
	Question string            `json:"question"`
	Choices  []QuestionChoices `json:"choices"`
}

type FormatAnswer struct {
	Answer string `json:"answer"`
	Code   string `json:"code"`
}

type Format struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required"`
}

type OllamaRequest struct {
	Model  string `json:"model"`
	System string `json:"system"`
	Prompt string `json:"prompt"`
	Format Format `json:"format"`
	Stream bool   `json:"stream"`
}

var PromptTemplate = OllamaRequest{
	Model:  "llama3.2:1b",
	System: "You're a a vocabulary teacher, that answers my questions about vocabulary. You not modify the question, and will keep the question and answers as received. You will respond with just the answer, and the respective code for that answer. Every answer MUST contain a code, from the choices provided. Example Input: context: 'Nearly 150 years later, the battle, which has been scrutinized by historians and immortalized in popular culture, is still steeped in controversy.'\n question: 'In the sentence above, immortalized has the same or almost the same meaning as:'\n choices:[{'key':'njnqx9','value':'reconnoitered'},{'key':'kps3g9','value':'disseminated'},{'key':'39ri9j','value':'circumvented'},{'key':'mo1y8u','value':'commemorated'}]\n answer: {answer:'commemorated', code:'mo1y8u'}}",
	Prompt: "",
	Format: Format{
		Type: "object",
		Properties: map[string]interface{}{
			"question": map[string]string{"type": "string"},
			"answer": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"answer": map[string]string{"type": "string"},
					"code":   map[string]string{"type": "string"},
				},
				"required": []string{"answer", "code"},
			},
		},
		Required: []string{"question", "answer"},
	},
	Stream: false,
}
