package billing

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"appengine"
	"appengine/blobstore"
	"appengine/datastore"
	"appengine/user"

	"github.com/gorilla/mux"
)

func authOnly(next myHandler) myHandler {
	return func(ctx *Context, w http.ResponseWriter, r *http.Request) error {
		if ctx.user == nil {
			url, err := user.LoginURL(ctx.c, "/login")
			if err != nil {
				return err
			}
			ctx.SetTitle("Please log in to continue ...")
			return ctx.Render(signinTmpl, url)
		}

		err := ctx.LoadUserSession()
		if err != nil {
			return err
		}

		if ctx.userSession == nil {
			url, err := user.LoginURL(ctx.c, "/login")
			if err != nil {
				return err
			}
			ctx.SetTitle("Invalid user ...")
			return ctx.Render(signinTmpl, url)
		}

		return next(ctx, w, r)
	}
}

func handleRoot(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	uploadURL, err := blobstore.UploadURL(ctx.c, "/upload", nil)
	if err != nil {
		return err
	}

	ctx.SetTitle("Upload New Bill")

	render := struct {
		UploadURL *url.URL
	}{uploadURL}
	return ctx.Render(uploadTmpl, render)
}

func handleView(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	s := r.FormValue("id")
	key, err := datastore.DecodeKey(s)

	if err != nil {
		return fmt.Errorf("Invalid key after decoding: %s", err.Error())
	}

	bill := new(Bill)

	err = datastore.Get(ctx.c, key, bill)

	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return ctx.NotFound()
		} else {
			return fmt.Errorf("Error on database get: %s", err.Error())
		}

	}

	ctx.Render(viewTmpl, bill)
	return nil
}

func handleDownload(ctx *Context, w http.ResponseWriter, r *http.Request) error {

	blobKey := appengine.BlobKey(r.FormValue("id"))
	stat, err := blobstore.Stat(ctx.c, blobKey)

	if err == datastore.ErrNoSuchEntity {
		return ctx.NotFound()
	}

	if err != nil {
		return err
	}

	hdr := w.Header()
	hdr.Set("Content-Disposition", "attachment; filename="+stat.Filename)
	hdr.Set("X-AppEngine-BlobKey", string(blobKey))
	return nil
}

func handleUpload(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	if ctx.user == nil {
		return fmt.Errorf("User cannot be nil!")
	}

	blobs, fields, err := blobstore.ParseUpload(ctx.r)
	if err != nil {
		return err
	}

	//vendor := getFormFieldString(fields, "vendor")
	amt := getFormFieldInt(fields, "amount")

	file := blobs["file"]
	if len(file) == 0 {
		ctx.c.Errorf("no file uploaded")
		http.Redirect(w, r, "/", http.StatusFound)
		return nil
	}

	b := Bill{
		Amt:      amt,
		PostedOn: time.Now(),
		PostedBy: ctx.user.String(),
		BlobKey:  file[0].BlobKey,
	}

	key := datastore.NewIncompleteKey(ctx.c, "Bill", nil)
	billKey, err := datastore.Put(ctx.c, key, &b)

	if err != nil {
		return err
	}

	http.Redirect(w, r, "/view/?id="+billKey.Encode(), http.StatusFound)

	return nil
}

func handleBillsDashboard(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	return ctx.Render(billsDashTmpl, "Hello Bills")
}

var (
	signinTmpl    = tmpl("signin.html")
	uploadTmpl    = tmpl("upload.html")
	viewTmpl      = tmpl("view.html")
	billsDashTmpl = tmpl("bills/dashboard.html")
)

func init() {
	r := mux.NewRouter()

	r.Handle("/", authOnly(handleRoot))

	setupAdminRoutes(r)
	setupLoginRoutes(r)

	r.Handle("/bills/dashboard", authOnly(handleBillsDashboard))
	r.Handle("/bills/view", authOnly(handleView))
	r.Handle("/bills/upload", authOnly(handleUpload))
	r.Handle("/bills/download/", authOnly(handleDownload))

	http.Handle("/", r)
}
