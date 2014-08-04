package billing

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"time"

	"appengine"
	"appengine/blobstore"
	"appengine/datastore"
	"appengine/user"
)

type Bill struct {
	ID         string
	PostedBy   string
	PostedOn   time.Time
	Vendor     string
	BlobKey    string
	Amt        int
	Paid       bool
	Reconciled bool
}

type Greeting struct {
	Author  string
	Content string
	Date    time.Time
}

type GreetingPage struct {
	User       *user.User
	SignOutURL string
	Greetings  []Greeting
}

func authOnly(next myHandler) myHandler {
	return func(ctx *Context, w http.ResponseWriter, r *http.Request) error {
		if ctx.user == nil {
			url, _ := user.LoginURL(ctx.c, "/")
			ctx.SetTitle("Please log in to continue ...")
			return ctx.Render(signinTmpl, url)
		}

		return next(ctx, w, r)
	}
}

// guestbookKey returns the key used for all guestbook entries.
func defaultBillCompanyKey(c appengine.Context) *datastore.Key {
	// The string "default_guestbook" here could be varied to have multiple guestbooks.
	return datastore.NewKey(c, "Bill", "default_parent_company", 0, nil)
}

func serveError(c appengine.Context, w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/plain")
	io.WriteString(w, "Internal Server Error")
	c.Errorf("%v", err)
}

func handleRoot(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	uploadURL, err := blobstore.UploadURL(ctx.c, "/upload", nil)
	if err != nil {
		return err
	}

	q := datastore.NewQuery("Bill").Ancestor(defaultBillCompanyKey(ctx.c)).Order("-PostedOn").Limit(10)
	bills := make([]Bill, 0, 10)
	keys, err := q.GetAll(ctx.c, &bills)
	if err != nil {
		return err
	}

	for idx, k := range keys {
		bills[idx].ID = k.Encode()
	}

	ctx.SetTitle("Upload New Bill")

	render := struct {
		UploadURL *url.URL
		Bills     []Bill
	}{uploadURL, bills}
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

type BlobInfo struct {
	BlobKey      appengine.BlobKey
	ContentType  string
	CreationTime *time.Time
	Filename     string // if provided
	Size         int64
}

const (
	blobInfoKind        = "__BlobInfo__"
	blobFileIndexKind   = "__BlobFileIndex__"
	blobKeyPropertyName = "blob_key"
	zeroKey             = appengine.BlobKey("")
)

func Stat(c appengine.Context, blobKey appengine.BlobKey) (*BlobInfo, error) {
	dskey := datastore.NewKey(c, blobInfoKind, string(blobKey), 0, nil)
	m := make(datastore.Map)
	if err := datastore.Get(c, dskey, m); err != nil {
		return nil, err
	}
	contentType, ok0 := m["content_type"].(string)
	filename, ok1 := m["filename"].(string)
	size, ok2 := m["size"].(int64)
	creation, ok3 := m["creation"].(datastore.Time)
	if !ok0 || !ok1 || !ok2 || !ok3 {
		return nil, errors.New("blobstore: invalid blob info")
	}
	bi := &BlobInfo{
		BlobKey:      blobKey,
		ContentType:  contentType,
		Filename:     filename,
		Size:         size,
		CreationTime: creation.Time(),
	}
	return bi, nil
}

func handleDownload(ctx *Context, w http.ResponseWriter, r *http.Request) error {

	blobKey := appengine.BlobKey(r.FormValue("id"))
	stat, err := Stat(ctx.c, blobKey)
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

	vendor := getFormFieldString(fields, "vendor")
	amt := getFormFieldInt(fields, "amount")

	file := blobs["file"]
	if len(file) == 0 {
		ctx.c.Errorf("no file uploaded")
		http.Redirect(w, r, "/", http.StatusFound)
		return nil
	}

	b := Bill{
		Amt:      amt,
		Vendor:   vendor,
		PostedOn: time.Now(),
		PostedBy: ctx.user.String(),
		BlobKey:  string(file[0].BlobKey),
	}

	key := datastore.NewIncompleteKey(ctx.c, "Bill", defaultBillCompanyKey(ctx.c))
	billKey, err := datastore.Put(ctx.c, key, &b)

	if err != nil {
		return err
	}

	http.Redirect(w, r, "/view/?id="+billKey.Encode(), http.StatusFound)

	return nil
}

var signinTmpl *template.Template = template.Must(template.ParseFiles("./tmpl/layout.html", "./tmpl/signin.html"))
var uploadTmpl *template.Template = template.Must(template.ParseFiles("./tmpl/layout.html", "./tmpl/upload.html"))
var viewTmpl *template.Template = template.Must(template.ParseFiles("./tmpl/layout.html", "./tmpl/view.html"))

func init() {
	http.Handle("/", authOnly(handleRoot))
	http.Handle("/view/", authOnly(handleView))
	http.Handle("/upload", authOnly(handleUpload))
	http.Handle("/download/", authOnly(handleDownload))
}
