package billing

import (
	"fmt"
	"html/template"
	"net/http"

	"appengine"
	"appengine/datastore"
	"appengine/user"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("gae-billing-system-secret-cookie-should-be-set-to-random"))

type UserSession struct {
	User    *User
	Company *Company
}

type UserPage struct {
	Session   *UserSession
	LogoutURL string
	User      *UserInfo
	Flashes   []interface{}
	Title     string
	Content   interface{}
	Path      string
}

type UserInfo struct {
	Admin     bool
	Name      string
	LogoutURL string
}

type myHandler func(context *Context, w http.ResponseWriter, r *http.Request) error

func (h myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	c := NewContext(w, r)
	err = h(c, w, r)
	if err != nil {
		c.HandleError(err)
		return
	}
}

func (c *Context) HandleError(err error) {
	if err == datastore.ErrNoSuchEntity {
		c.NotFound()
		return
	}

	serveError(c.c, c.w, err)
	return
}

type Context struct {
	title       string
	user        *user.User
	userSession *UserSession
	session     *sessions.Session
	admin       bool
	c           appengine.Context
	w           http.ResponseWriter
	r           *http.Request
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	c := &Context{
		w: w,
		r: r,
	}

	c.c = appengine.NewContext(r)
	c.user = user.Current(c.c)
	c.admin = user.IsAdmin(c.c)

	return c
}

func (c *Context) Session() *sessions.Session {

	if c.session == nil {
		session, _ := store.Get(c.r, "gae-billing-session")
		c.session = session
	}

	return c.session
}

func (c *Context) Flash(format string, args ...interface{}) {
	c.Session().AddFlash(fmt.Sprintf(format, args...))
}

func (c *Context) SaveSession() error {
	if c.session == nil {
		return nil
	}

	return c.session.Save(c.r, c.w)
}

func (c *Context) Render(t *template.Template, content interface{}) error {
	p := UserPage{
		Content: content,
		Flashes: c.Session().Flashes(),
		Path:    c.r.URL.Path,
	}

	if c.user != nil {

		url, err := user.LogoutURL(c.c, "/")
		if err != nil {
			return err
		}

		p.User = &UserInfo{
			Admin:     c.admin,
			Name:      c.user.String(),
			LogoutURL: url,
		}

		if c.session != nil {
			p.Session = c.userSession
		}
	}

	if c.title != "" {
		p.Title = c.title
	}

	err := c.SaveSession()
	if err != nil {
		return err
	}

	c.w.Header().Set("Content-Type", "text/html")

	err = t.Execute(c.w, p)

	return err
}

func (ctx *Context) LoadUserSession() error {
	if ctx.user == nil {
		return nil
	}

	email := ctx.user.String()

	billUser, billCompany, err := ctx.GetUserAndCompanyByEmail(email)

	if err != nil {
		return err
	}

	ctx.Debugf("User session: %v", billUser)
	ctx.Debugf("Company session: %v", billCompany)

	if billUser != nil && billCompany != nil {
		ctx.userSession = &UserSession{billUser, billCompany}
	}

	return nil
}

func (ctx *Context) Redirect(url string) error {
	err := ctx.SaveSession()
	if err != nil {
		return err
	}

	http.Redirect(ctx.w, ctx.r, url, http.StatusFound)
	return nil
}

func (ctx *Context) Debugf(format string, args ...interface{}) {
	ctx.c.Debugf(format, args...)
}

func (c *Context) NotFound() error {
	http.NotFound(c.w, c.r)
	return nil
}

func (c *Context) SetTitle(t string) {
	c.title = t
}
