package model

// RespAnswer model
type RespAnswer struct {
	QuestionID int `rethinkdb:"question_id"`
	AnswerID   int `rethinkdb:"answer_id"`
	// Answer Answer `rethinkdb:"answer_id ,reference" rethinkdb_ref:"answer_id"`
}

// RespSurvey model
type RespSurvey struct {
	ID         string       `rethinkdb:"survey_id"`
	AnswerList []RespAnswer `rethinkdb:"answer_list"`
}

// Respondent model
type Respondent struct {
	ID string `rethinkdb:"id,omitempty"`
	// Name    string `rethinkdb:"name"`
	SurveyList []RespSurvey `rethinkdb:"survey_list"`
}
