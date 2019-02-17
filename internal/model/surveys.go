package model

// Answer model
type Answer struct {
	ID      int    `rethinkdb:"answer_id" json:"answer_id"`
	Text    string `rethinkdb:"text" json:"text"`
	Counter int    `rethinkdb:"counter" json:"counter"`
}

// Question model
type Question struct {
	ID              int      `rethinkdb:"question_id" json:"question_id"`
	Title           string   `rethinkdb:"question_title" json:"question_title,omitempty"`
	AnswerList      []Answer `rethinkdb:"answer_list" json:"answer_list"`
	ExclusiveAnswer bool     `rethinkdb:"exclusive" json:"exclusive"`
}

// Survey model
type Survey struct {
	ID           string     `rethinkdb:"id,omitempty" json:"id"`
	Title        string     `rethinkdb:"survey_title" json:"survey_title,omitempty"`
	QuestionList []Question `rethinkdb:"question_list" json:"question_list"`
}
