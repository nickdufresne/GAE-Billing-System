package billing

import (
	"appengine/user"
	"net/http"

	"github.com/gorilla/mux"
)

func handleLogin(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	if user.IsAdmin(ctx.c) {
		ctx.Flash("Welcome admin user!")
		ctx.Redirect("/admin/dashboard")
		return nil
	}

	ctx.Flash("Welcome regular user!")
	ctx.Redirect("/bills/dashboard")
	return nil
}

func setupLoginRoutes(router *mux.Router) {
	router.Handle("/login", myHandler(handleLogin))
}
