package billing

import (
	"html/template"
	"path/filepath"
	"time"
)

var tmplFuncMap = template.FuncMap{
	"date": tmplDate,
	"time": tmplTime,
}

func tmpl(p string) *template.Template {
	path := filepath.Join("./tmpl", p)
	return template.Must(template.New("layout.html").Funcs(tmplFuncMap).ParseFiles("./tmpl/layout.html", path))
}

func tmplDate(date time.Time) string {
	if date.IsZero() {
		return ""
	}

	return date.Format("01/02/06")
}

func tmplTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.Format("01/02/06 03:04:05pm")
}
