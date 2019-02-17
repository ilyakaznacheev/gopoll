package poll

import "html/template"

// ContextDisplayPollAnswer a poll template answer context
type ContextDisplayPollAnswer struct {
	Num    string
	Answer string
}

// ContextDisplayPollQuestion is a poll template question context
type ContextDisplayPollQuestion struct {
	Title      string
	Subtitle   string
	Exclusive  bool
	Num        string
	AnswerList []ContextDisplayPollAnswer
}

// ContextDisplayPoll is a poll template context
type ContextDisplayPoll struct {
	Link         template.URL
	Title        string
	QuestionList []ContextDisplayPollQuestion
}

// PollVote is a poll vote data
type PollVote []struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ContextDisplayPollElement is a poll list element context
type ContextDisplayPollElement struct {
	Title         string
	Link          template.URL
	SurveyLink    template.URL
	QuestionCount int
}

// ContextDisplayPollList is a poll list context
type ContextDisplayPollList struct {
	List []ContextDisplayPollElement
}

// ContextPollChartDataset is a poll chart dataset info
type ContextPollChartDataset struct {
	Label           string   `json:"label"`
	Data            []int    `json:"data"`
	BackgroundColor []string `json:"backgroundColor"`
	BorderColor     []string `json:"borderColor"`
	BorderWidth     int      `json:"borderWidth"`
}

// ContextPollChart is a poll chart settings
type ContextPollChart struct {
	Labels   []string                  `json:"labels"`
	Datasets []ContextPollChartDataset `json:"datasets"`
}

// ContextPollChartSet set fof data for one chart
type ContextPollChartSet struct {
	ID   int              `json:"id"`
	Data ContextPollChart `json:"data"`
}

// ContextPollPage set of data for charts
type ContextPollPage struct {
	DataSet []ContextPollChartSet `json:"dataset"`
}

// ContextChartMainElement is a chart
type ContextChartMainElement struct {
	ID         int
	ChartTitle string
}

// ContextChartMain is a list of polls
type ContextChartMain struct {
	Element []ContextChartMainElement
	Title   string
	Link    template.URL
}

// ContextCreatePollAnswer questions for creqtion
type ContextCreatePollAnswer struct {
	Answers []int
}

// ContextCreatePollQuestion ia s contecst for poll creation fith fixed numbers of questions and answers
type ContextCreatePollQuestion struct {
	Questions []ContextCreatePollAnswer
}
