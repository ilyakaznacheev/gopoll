package poll

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/DronRathore/goexpress"

	r "gopkg.in/rethinkdb/rethinkdb-go.v5"
)

// Run starts an app
func Run(port, confPath, static string) error {
	conf, err := loadConfig(confPath)
	if err != nil {
		return err
	}

	rs, err := initDB(conf)
	if err != nil {
		return err
	}

	var app = goexpress.Express()

	rh := NewRequestHandler(rs, *conf, static)

	a := auth{
		user:     conf.Admin.Name,
		password: conf.Admin.Pass,
	}

	app.Get("/admin", a.check(rh.HandleAdminPollList))
	app.Get("/admin/create", a.check(rh.HandleAdminCreatePoll))
	app.Post("/admin/create", a.check(rh.HandleAdminSavePoll))
	app.Get("/admin/survey/:id", a.check(rh.HandleAdminReviewPoll))
	app.Get("/chart/:id", a.check(rh.HandleChartData))
	app.Get("/survey/:id", rh.HandlePoll)
	app.Post("/vote/:id", rh.HandlePollVote)
	app.Get("/thanks", rh.HandleThanks)

	app.BeforeShutdown(func(e goexpress.ExpressInterface) {
		rs.Close()
	})

	fmt.Println("Starting server")
	app.Start(port)
	return nil
}

func initDB(conf *Config) (*r.Session, error) {
	conn := r.ConnectOpts{
		Address: conf.DB.Address,
	}
	if conf.DB.Username != nil && conf.DB.Password != nil {
		conn.Database = conf.DB.Name
		conn.Username = *conf.DB.Username
		conn.Password = *conf.DB.Password
	}

	rs, err := r.Connect(conn)
	if err != nil {
		return nil, err
	}

	r.DB(conf.DB.Name).TableCreate(tabSurveys, r.TableCreateOpts{PrimaryKey: "id"}).Run(rs)
	r.DB(conf.DB.Name).TableCreate(tabRespondents, r.TableCreateOpts{PrimaryKey: "id"}).Run(rs)

	return rs, nil
}

// auth is authorization manager
type auth struct {
	user     string
	password string
}

// check authorization
func (a *auth) check(mw goexpress.Middleware) goexpress.Middleware {
	return func(req goexpress.Request, res goexpress.Response) {
		res.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		s := strings.SplitN(req.Header().Get("authorization"), " ", 2)
		if len(s) != 2 {
			res.Error(http.StatusUnauthorized, "Not authorized")
			return
		}

		b, err := base64.StdEncoding.DecodeString(s[1])
		if err != nil {
			res.Error(http.StatusUnauthorized, err.Error())
			return
		}

		pair := strings.SplitN(string(b), ":", 2)
		if len(pair) != 2 {
			res.Error(http.StatusUnauthorized, "Not authorized")
			return
		}

		if pair[0] != a.user || pair[1] != a.password {
			res.Error(http.StatusUnauthorized, "Not authorized")
			return
		}

		mw(req, res)
	}
}
