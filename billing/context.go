package billing

import (
	"html/template"
	"net/http"

	"appengine"
	"appengine/user"
)

type UserPage struct {
	LogoutURL string
	User      *UserInfo
	Title     string
	Content   interface{}
}

type UserInfo struct {
	Name      string
	LogoutURL string
}

type myHandler func(context *Context, w http.ResponseWriter, r *http.Request) error

func (h myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := NewContext(w, r)
	err := h(c, w, r)
	if err != nil {
		serveError(c.c, w, err)
	}
}

type Context struct {
	title string
	user  *user.User
	c     appengine.Context
	w     http.ResponseWriter
	r     *http.Request
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	c := &Context{
		w: w,
		r: r,
	}

	c.c = appengine.NewContext(r)
	c.user = user.Current(c.c)

	return c
}

func (c *Context) Render(t *template.Template, content interface{}) error {
	p := UserPage{
		Content: content,
	}

	if c.user != nil {

		url, err := user.LogoutURL(c.c, "/")
		if err != nil {
			return err
		}
		p.User = &UserInfo{
			Name:      c.user.String(),
			LogoutURL: url,
		}
	}

	if c.title != "" {
		p.Title = c.title
	}

	c.w.Header().Set("Content-Type", "text/html")

	err := t.Execute(c.w, p)

	return err
}

func (c *Context) NotFound() error {
	http.NotFound(c.w, c.r)
	return nil
}

func (c *Context) SetTitle(t string) {
	c.title = t
}
