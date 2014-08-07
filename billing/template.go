package billing

import (
	"fmt"
	"html/template"
	"path/filepath"
	"time"
)

var tmplFuncMap = template.FuncMap{
	"date":  tmplDate,
	"time":  tmplTime,
	"money": tmplMoney,
}

var tmplAdminFuncMap = template.FuncMap{
	"date":                 tmplDate,
	"time":                 tmplTime,
	"sidebarLink":          tmplSidebarLink,
	"sidebarLinkWithCount": tmplSidebarLinkWithCount,
	"money":                tmplMoney,
}

func adminTmpl(p string) *template.Template {
	path := filepath.Join("./tmpl/admin", p)
	templates := []string{
		"./tmpl/admin/layout.html",
		path,
		"./tmpl/_user_menu.html",
		"./tmpl/admin/_menu.html",
	}
	return template.Must(template.New("layout.html").Funcs(tmplAdminFuncMap).ParseFiles(templates...))
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

func tmplSidebarLink(path string, name string, currentPath string) template.HTML {
	if currentPath == path {
		return template.HTML(fmt.Sprintf("<li class=\"active\"><a href=\"%s\"> %s </a></li>", path, name))
	}

	return template.HTML(fmt.Sprintf("<li><a href=\"%s\"> %s </a></li>", path, name))
}

func tmplSidebarLinkWithCount(path string, name string, count int, currentPath string) template.HTML {
	if currentPath == path {
		return template.HTML(fmt.Sprintf("<li class=\"active\"><a href=\"%s\"> %s (%d) </a></li>", path, name, count))
	}

	return template.HTML(fmt.Sprintf("<li><a href=\"%s\"> %s (%d) </a></li>", path, name, count))
}

func tmplMoney(v int) string {
	f := float64(v) / 100.0
	return fmt.Sprintf("%0.2f", f)
}
