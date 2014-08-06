package billing

import (
	"net/http"
	"time"

	"appengine/datastore"
	"appengine/user"

	"github.com/gorilla/mux"
)

func adminOnly(next myHandler) myHandler {
	return func(ctx *Context, w http.ResponseWriter, r *http.Request) error {
		if !user.IsAdmin(ctx.c) {
			http.Redirect(w, r, "/bills/dashboard", http.StatusFound)
			return nil
		}

		return next(ctx, w, r)
	}
}

type DashboardPage struct {
	Users     []*User
	Companies []*Company
	Vendors   []*Vendor
}

func handleAdminDashboard(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	companies, err := ctx.GetAllCompanies()
	if err != nil {
		return err
	}

	users, err := ctx.GetAllUsers()
	if err != nil {
		return err
	}

	vendors, err := ctx.GetAllVendors()
	if err != nil {
		return err
	}

	// associate companies with users ...
	// if there are more than 10 companies this wont work because we limit our companies to 10 in GetAll
	for _, user := range users {
		for _, c := range companies {
			if user.CompanyKey.Encode() == c.ID {
				user.Company = c
			}
		}
	}

	// associate companies with vendors ...
	// if there are more than 10 companies this wont work because we limit our companies to 10 in GetAll
	for _, vendor := range vendors {
		for _, c := range companies {
			if vendor.CompanyKey.Encode() == c.ID {
				vendor.Company = c
			}
		}
	}

	return ctx.Render(adminDashTmpl, DashboardPage{users, companies, vendors})
}

func handleNewCompany(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	return ctx.Render(newCompanyTmpl, nil)
}

func handleCreateCompany(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	name := r.FormValue("name")
	comp := Company{
		Name:      name,
		CreatedOn: time.Now(),
		CreatedBy: ctx.user.String(),
	}

	key := datastore.NewIncompleteKey(ctx.c, "Company", defaultCompanyKey(ctx.c))
	_, err := datastore.Put(ctx.c, key, &comp)

	if err != nil {
		return err
	}

	http.Redirect(w, r, "/admin/dashboard", http.StatusFound)
	return nil
}

type NewUserForm struct {
	User           *User
	ValidationErrs []string
	Companies      []*Company
}

func handleCreateUser(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	var companyKey *datastore.Key
	var err error

	vErrs := []string{}

	email := r.FormValue("email")

	if email == "" {
		vErrs = append(vErrs, "Email must be valid")
	} else {
		found, err := ctx.UserEmailExists(email)
		if err != nil {
			return err
		}

		if found {
			vErrs = append(vErrs, "User already exists")
		}
	}

	companyID := r.FormValue("company")

	if companyID == "" {
		vErrs = append(vErrs, "You must select a company")
	} else {
		companyKey, err = datastore.DecodeKey(companyID)
		if err != nil {
			vErrs = append(vErrs, "Invalid company selected")
		}
	}

	user := User{
		CompanyKey: companyKey,
		Email:      email,
		CreatedOn:  time.Now(),
		CreatedBy:  ctx.user.String(),
	}

	if len(vErrs) > 0 {
		companies, err := ctx.GetAllCompanies()
		if err != nil {
			return err
		}
		ctx.Render(newUserTmpl, NewUserForm{&user, vErrs, companies})
		return nil
	}

	key := datastore.NewIncompleteKey(ctx.c, "User", companyKey)
	_, err = datastore.Put(ctx.c, key, &user)

	if err != nil {
		return err
	}

	http.Redirect(w, r, "/admin/dashboard", http.StatusFound)
	return nil
}

func handleNewUser(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	companies, err := ctx.GetAllCompanies()
	if err != nil {
		return err
	}
	return ctx.Render(newUserTmpl, NewUserForm{&User{}, []string{}, companies})
}

type NewVendorForm struct {
	Vendor         *Vendor
	ValidationErrs []string
	Companies      []*Company
}

func handleCreateVendor(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	var companyKey *datastore.Key
	var err error

	vErrs := []string{}

	name := r.FormValue("name")

	if name == "" {
		vErrs = append(vErrs, "Name must be valid")
	}

	companyID := r.FormValue("company")

	if companyID == "" {
		vErrs = append(vErrs, "You must select a company")
	} else {
		companyKey, err = datastore.DecodeKey(companyID)
		if err != nil {
			vErrs = append(vErrs, "Invalid company selected")
		}
	}

	vendor := Vendor{
		CompanyKey: companyKey,
		Name:       name,
		CreatedOn:  time.Now(),
		CreatedBy:  ctx.user.String(),
	}

	if len(vErrs) > 0 {
		companies, err := ctx.GetAllCompanies()
		if err != nil {
			return err
		}
		ctx.Render(newVendorTmpl, NewVendorForm{&vendor, vErrs, companies})
		return nil
	}

	key := datastore.NewIncompleteKey(ctx.c, "Vendor", companyKey)
	_, err = datastore.Put(ctx.c, key, &vendor)

	if err != nil {
		return err
	}

	return ctx.Redirect("/admin/dashboard")
}

func handleNewVendor(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	companies, err := ctx.GetAllCompanies()
	if err != nil {
		return err
	}
	return ctx.Render(newVendorTmpl, NewVendorForm{&Vendor{}, []string{}, companies})
}

func handleViewCompany(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	id := r.FormValue("id")
	c, err := ctx.GetCompanyByID(id)
	if err != nil {
		return err
	}

	users, err := ctx.GetCompanyUsers(c)
	if err != nil {
		return err
	}

	vendors, err := ctx.GetCompanyVendors(c)
	if err != nil {
		return err
	}

	c.Users = users
	c.Vendors = vendors

	return ctx.Render(viewCompanyTmpl, c)
}

func handleViewVendor(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	return ctx.Render(viewVendorTmpl, nil)
}

func handleDeleteUser(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	id := r.FormValue("id")

	u, err := ctx.GetUserByID(id)
	if err != nil {
		return err
	}

	err = ctx.DeleteUser(id)

	if err != nil {
		return err
	}

	ctx.Flash("User %s deleted!", u.Email)

	return ctx.Redirect("/admin/dashboard")
}

var (
	adminDashTmpl   = tmpl("admin/dashboard.html")
	newCompanyTmpl  = tmpl("admin/new_company.html")
	newUserTmpl     = tmpl("admin/new_user.html")
	viewCompanyTmpl = tmpl("admin/view_company.html")
	newVendorTmpl   = tmpl("admin/new_vendor.html")
	viewVendorTmpl  = tmpl("admin/view_vendor.html")
)

func setupAdminRoutes(router *mux.Router) {
	router.Handle("/admin/dashboard", adminOnly(handleAdminDashboard))
	router.Handle("/admin/user/new", adminOnly(handleNewUser))
	router.Handle("/admin/user/create", adminOnly(handleCreateUser))
	router.Handle("/admin/user/delete", adminOnly(handleDeleteUser))
	router.Handle("/admin/company/new", adminOnly(handleNewCompany))
	router.Handle("/admin/company/create", adminOnly(handleCreateCompany))
	router.Handle("/admin/company/view", adminOnly(handleViewCompany))

	router.Handle("/admin/vendor/new", adminOnly(handleNewVendor))
	router.Handle("/admin/vendor/create", adminOnly(handleCreateVendor))
	router.Handle("/admin/vendor/view", adminOnly(handleViewVendor))

}
