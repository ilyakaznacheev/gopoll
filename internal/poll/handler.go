package poll

import (
	"bytes"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/DronRathore/goexpress"
	"github.com/ilyakaznacheev/gopoll/internal/model"
	r "gopkg.in/rethinkdb/rethinkdb-go.v5"
)

const (
	tabSurveys     = "surveys"
	tabRespondents = "respondents"
	userCookieName = "gopoll_respondentname"
)

// RequestHandler handles HTTP requests
type RequestHandler struct {
	rs       *r.Session
	conf     Config
	template string
}

// NewRequestHandler create new handler instance
func NewRequestHandler(rs *r.Session, conf Config, static string) *RequestHandler {
	rand.Seed(time.Now().UTC().UnixNano())
	return &RequestHandler{
		rs:       rs,
		conf:     conf,
		template: path.Join(static, "templates") + "/",
	}
}

// HandleAdminCreatePoll creates new poll
func (h *RequestHandler) HandleAdminCreatePoll(req goexpress.Request, res goexpress.Response) {
	m := ContextCreatePollQuestion{
		Questions: make([]ContextCreatePollAnswer, h.conf.Poll.Questions),
	}
	for q := 0; q < h.conf.Poll.Questions; q++ {
		m.Questions[q].Answers = make([]int, h.conf.Poll.Answers)
		for a := 0; a < h.conf.Poll.Answers; a++ {
			m.Questions[q].Answers[a] = a
		}
	}

	tmpl := template.Must(template.ParseFiles(h.template + "create.html"))

	var tpl bytes.Buffer
	err := tmpl.Execute(&tpl, m)
	if err != nil {
		res.Error(http.StatusInternalServerError, err.Error())
	}

	res.WriteBytes(tpl.Bytes())
}

// HandleAdminSavePoll save new pall
func (h *RequestHandler) HandleAdminSavePoll(req goexpress.Request, res goexpress.Response) {
	var vote PollVote
	req.JSON().Decode(&vote)

	answers := make(map[int][]string)
	questions := make(map[int]string)
	exclusion := make(map[int]bool)

	poll := model.Survey{}
	for _, v := range vote {
		key := strings.Split(v.Name, ":")[0]

		switch key {
		case "survey":
			poll.Title = v.Value
		case "question":
			keyInt, _ := strconv.Atoi(key)
			questions[keyInt] = v.Value
		case "excl":
			keyInt, _ := strconv.Atoi(key)
			b, _ := strconv.ParseBool(v.Value)
			exclusion[keyInt] = !b
		default:
			if v.Value != "" {
				keyInt, _ := strconv.Atoi(key)
				answers[keyInt] = append(answers[keyInt], v.Value)
			}
		}
	}

	poll.QuestionList = make([]model.Question, 0, len(answers))
	i := 0
	for key, val := range answers {
		a := make([]model.Answer, 0, len(val))
		for aidx, ans := range val {
			a = append(a, model.Answer{
				ID:   aidx,
				Text: ans,
			})
		}
		qKey, _ := questions[key]
		eKey, _ := exclusion[key]

		q := model.Question{
			ID:              i,
			Title:           qKey,
			ExclusiveAnswer: eKey,
			AnswerList:      a,
		}
		poll.QuestionList = append(poll.QuestionList, q)
		i++

	}

	if poll.Title == "" || len(poll.QuestionList) == 0 {
		res.Error(http.StatusBadRequest, "wrong poll content")
	}

	_, err := r.DB(h.conf.DB.Name).Table(tabSurveys).Insert(&poll, r.InsertOpts{ReturnChanges: true}).RunWrite(h.rs)
	if err != nil {
		res.Error(http.StatusInternalServerError, err.Error())
	}

	res.Write("poll created")
}

// HandleAdminPollList returns a poll list
func (h *RequestHandler) HandleAdminPollList(req goexpress.Request, res goexpress.Response) {
	result, err := r.DB(h.conf.DB.Name).Table(tabSurveys).Run(h.rs)
	if err != nil {
		res.Error(http.StatusInternalServerError, err.Error())
	}

	var rowList []model.Survey
	result.All(&rowList)

	ctx := ContextDisplayPollList{make([]ContextDisplayPollElement, 0, len(rowList))}
	for _, s := range rowList {
		ctx.List = append(ctx.List, ContextDisplayPollElement{
			Title:         s.Title,
			Link:          template.URL("/admin/survey/" + s.ID),
			SurveyLink:    template.URL("/survey/" + s.ID),
			QuestionCount: len(s.QuestionList),
		})
	}

	tmpl := template.Must(template.ParseFiles(h.template + "poll_list.html"))

	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, ctx)
	if err != nil {
		res.Error(http.StatusInternalServerError, err.Error())
	}

	res.WriteBytes(tpl.Bytes())
}

// HandlePoll returns a poll view
func (h *RequestHandler) HandlePoll(req goexpress.Request, res goexpress.Response) {
	type context struct {
		model.Survey
		Link template.URL
	}
	pollID := req.Params().Get("id")

	result, err := r.DB(h.conf.DB.Name).Table(tabSurveys).Get(pollID).Run(h.rs)
	if err != nil {
		res.Error(http.StatusInternalServerError, err.Error())
	}
	if result.IsNil() {
		res.Error(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	var poll model.Survey
	result.One(&poll)

	ctx := ContextDisplayPoll{
		Link:         template.URL("/vote/" + poll.ID),
		Title:        poll.Title,
		QuestionList: make([]ContextDisplayPollQuestion, 0, len(poll.QuestionList)),
	}

	for idxQ, q := range poll.QuestionList {
		answers := make([]ContextDisplayPollAnswer, 0, len(q.AnswerList))
		for _, a := range q.AnswerList {
			answers = append(answers, ContextDisplayPollAnswer{
				Num:    strconv.Itoa(a.ID),
				Answer: a.Text,
			})
		}
		ctx.QuestionList = append(ctx.QuestionList, ContextDisplayPollQuestion{
			Title:      q.Title,
			Subtitle:   fmt.Sprintf("Question %d", idxQ+1),
			Exclusive:  q.ExclusiveAnswer,
			AnswerList: answers,
			Num:        strconv.Itoa(q.ID),
		})
	}

	tmpl := template.Must(template.ParseFiles(h.template + "poll.html"))

	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, ctx)
	if err != nil {
		res.Error(http.StatusInternalServerError, err.Error())
	}

	res.WriteBytes(tpl.Bytes())
}

// HandlePollVote saves vote results
func (h *RequestHandler) HandlePollVote(req goexpress.Request, res goexpress.Response) {
	userName := req.Cookie().Get(userCookieName)
	pollID := req.Params().Get("id")

	var respondent *model.Respondent

	if userName != "" {
		result, err := r.DB(h.conf.DB.Name).Table(tabRespondents).Get(userName).Run(h.rs)
		if err != nil {
			res.Error(http.StatusInternalServerError, err.Error())
		}

		if result.IsNil() {
			userName = ""
		} else {
			result.One(&respondent)
		}
	}

	if respondent == nil {
		respondent = &model.Respondent{}
		result, err := r.DB(h.conf.DB.Name).Table(tabRespondents).Insert(respondent, r.InsertOpts{ReturnChanges: true}).RunWrite(h.rs)
		if err != nil {
			res.Error(http.StatusInternalServerError, err.Error())
		}
		if result.Inserted > 0 {
			respondent.ID = result.GeneratedKeys[0]
		} else {
			res.Error(http.StatusInternalServerError, "Respondent creation error")
		}

	}

	// get poll data
	var poll model.Survey
	result, err := r.DB(h.conf.DB.Name).Table(tabSurveys).Get(pollID).Run(h.rs)
	if err != nil {
		res.Error(http.StatusInternalServerError, err.Error())
	}

	// parse json
	result.One(&poll)
	var vote PollVote
	req.JSON().Decode(&vote)

	// check if vote already done
	for _, s := range respondent.SurveyList {
		if s.ID == pollID {
			res.Error(http.StatusBadRequest, "vote already done")
			return
		}
	}

	answers := make([]model.RespAnswer, 0, len(vote))
	for _, v := range vote {
		idxQ, _ := strconv.Atoi(v.Name)
		idxA, _ := strconv.Atoi(v.Value)

		answers = append(answers, model.RespAnswer{
			QuestionID: idxQ,
			AnswerID:   idxA,
		})

		for idxPQ, q := range poll.QuestionList {
			if q.ID == idxQ {
				for idxPA, a := range q.AnswerList {
					if a.ID == idxA {
						// add vote counter
						poll.QuestionList[idxPQ].AnswerList[idxPA].Counter++
					}
				}
			}
		}
	}

	respondent.SurveyList = append(respondent.SurveyList, model.RespSurvey{
		ID:         pollID,
		AnswerList: answers,
	})

	// save respondent data
	_, err = r.DB(h.conf.DB.Name).Table(tabRespondents).Update(respondent).Run(h.rs)
	if err != nil {
		res.Error(http.StatusInternalServerError, err.Error())
	}

	// save poll data
	_, err = r.DB(h.conf.DB.Name).Table(tabSurveys).Update(poll).Run(h.rs)
	if err != nil {
		res.Error(http.StatusInternalServerError, err.Error())
	}

	res.Cookie().Add(&http.Cookie{
		Name:     userCookieName,
		Value:    respondent.ID,
		Path:     "/",
		HttpOnly: true,
	})
	res.JSON(vote)
}

// HandleAdminReviewPoll returns a chart template
func (h *RequestHandler) HandleAdminReviewPoll(req goexpress.Request, res goexpress.Response) {
	pollID := req.Params().Get("id")

	var poll model.Survey
	result, err := r.DB(h.conf.DB.Name).Table(tabSurveys).Get(pollID).Run(h.rs)
	if err != nil {
		res.Error(http.StatusInternalServerError, err.Error())
	}
	result.One(&poll)

	charts := ContextChartMain{
		Element: make([]ContextChartMainElement, 0, len(poll.QuestionList)),
		Title:   poll.Title,
		Link:    template.URL("/chart/" + pollID),
	}
	for _, q := range poll.QuestionList {
		charts.Element = append(charts.Element, ContextChartMainElement{
			ID:         q.ID,
			ChartTitle: q.Title,
		})
	}

	tmpl := template.Must(template.ParseFiles(h.template + "chart.html"))

	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, &charts)
	if err != nil {
		res.Error(http.StatusInternalServerError, err.Error())
	}

	res.WriteBytes(tpl.Bytes())
}

// HandleChartData returns chart dataset
func (h *RequestHandler) HandleChartData(req goexpress.Request, res goexpress.Response) {
	pollID := req.Params().Get("id")

	var poll model.Survey
	result, err := r.DB(h.conf.DB.Name).Table(tabSurveys).Get(pollID).Run(h.rs)
	if err != nil {
		res.Error(http.StatusInternalServerError, err.Error())
	}
	result.One(&poll)

	c := ContextPollPage{
		DataSet: make([]ContextPollChartSet, 0, len(poll.QuestionList)),
	}

	for _, q := range poll.QuestionList {
		labels := make([]string, 0, len(q.AnswerList))
		dataset := ContextPollChartDataset{
			Label:           "# of Votes",
			BorderWidth:     1,
			Data:            make([]int, 0, len(q.AnswerList)),
			BackgroundColor: make([]string, 0, len(q.AnswerList)),
		}
		// := make([]string,0, 1)
		for _, a := range q.AnswerList {
			labels = append(labels, a.Text)
			dataset.Data = append(dataset.Data, a.Counter)
			dataset.BackgroundColor = append(dataset.BackgroundColor,
				fmt.Sprintf("rgba(%d, %d, %d, 0.5)", rand.Intn(256), rand.Intn(256), rand.Intn(256)))
		}

		c.DataSet = append(c.DataSet, ContextPollChartSet{
			ID: q.ID,
			Data: ContextPollChart{
				Labels:   labels,
				Datasets: []ContextPollChartDataset{dataset},
			},
		})
	}

	res.JSON(&c)
}

// HandleThanks returns poll ending page
func (h *RequestHandler) HandleThanks(req goexpress.Request, res goexpress.Response) {
	tmpl := template.Must(template.ParseFiles(h.template + "thanks.html"))

	var tpl bytes.Buffer
	err := tmpl.Execute(&tpl, nil)
	if err != nil {
		res.Error(http.StatusInternalServerError, err.Error())
	}

	res.WriteBytes(tpl.Bytes())
}
