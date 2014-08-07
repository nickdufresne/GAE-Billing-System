package billing

import (
	"html/template"
	"net/http"
	"net/url"
	"time"

	"appengine/blobstore"
	"appengine/datastore"
	"appengine/user"

	"github.com/gorilla/mux"
)

type DashboardInfo struct {
	VendorCount  int
	UserCount    int
	BillCount    int
	CompanyCount int
	Path         string
}

type DashboardPage struct {
	Dashboard *DashboardInfo
	Page      interface{}
}

func adminOnly(next myHandler) myHandler {
	return func(ctx *Context, w http.ResponseWriter, r *http.Request) error {
		if !user.IsAdmin(ctx.c) {
			http.Redirect(w, r, "/bills/dashboard", http.StatusFound)
			return nil
		}

		return next(ctx, w, r)
	}
}

type DashboardHomePage struct {
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

	return ctx.renderAdmin(adminDashTmpl, DashboardHomePage{users, companies, vendors})
}

func handleNewCompany(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	ctx.Debugf("Handle New Company ...")
	return ctx.renderAdmin(newCompanyTmpl, nil)
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
		ctx.renderAdmin(newUserTmpl, NewUserForm{&user, vErrs, companies})
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
	return ctx.renderAdmin(newUserTmpl, NewUserForm{&User{}, []string{}, companies})
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
		ctx.renderAdmin(newVendorTmpl, NewVendorForm{&vendor, vErrs, companies})
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
	return ctx.renderAdmin(newVendorTmpl, NewVendorForm{&Vendor{}, []string{}, companies})
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

	return ctx.renderAdmin(viewCompanyTmpl, c)
}

func handleViewVendor(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	return ctx.renderAdmin(viewVendorTmpl, nil)
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

func (ctx *Context) GetDashboardInfo() (*DashboardInfo, error) {
	users, err := ctx.GetUserCount()
	if err != nil {
		return nil, err
	}
	companies, err := ctx.GetCompanyCount()
	if err != nil {
		return nil, err
	}
	vendors, err := ctx.GetVendorCount()
	if err != nil {
		return nil, err
	}
	bills, err := ctx.GetBillCount()
	if err != nil {
		return nil, err
	}
	di := &DashboardInfo{
		UserCount:    users,
		CompanyCount: companies,
		VendorCount:  vendors,
		BillCount:    bills,
		Path:         ctx.r.URL.Path,
	}

	return di, nil
}

func (ctx *Context) renderAdmin(t *template.Template, content interface{}) error {
	di, err := ctx.GetDashboardInfo()
	if err != nil {
		return err
	}

	return ctx.Render(t, DashboardPage{di, content})
}

func handleAdminCompanies(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	companies, err := ctx.GetAllCompanies()
	if err != nil {
		return err
	}
	return ctx.renderAdmin(viewCompanies, companies)
}

func handleAdminUsers(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	users, err := ctx.GetAllUsers()
	if err != nil {
		return err
	}

	err = ctx.LoadUserCompanies(users)
	if err != nil {
		return err
	}

	return ctx.renderAdmin(viewUsers, users)
}

func handleAdminVendors(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	vendors, err := ctx.GetAllVendors()
	if err != nil {
		return err
	}

	err = ctx.LoadVendorCompanies(vendors)
	if err != nil {
		return err
	}

	return ctx.renderAdmin(viewVendors, vendors)
}

func handleAdminBills(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	//ctx.GetAllBills().FetchVendor().FetchCompany().Slice()
	//tmpl("viewBills").Render(results)

	bills, err := ctx.GetAllBills()
	if err != nil {
		return err
	}

	err = ctx.LoadBillCompanies(bills)
	if err != nil {
		return err
	}

	ctx.LoadBillVendors(bills)
	if err != nil {
		return err
	}

	return ctx.renderAdmin(viewBills, bills)
}

type NewBillForm struct {
	Bill           *Bill
	ValidationErrs []string
	Vendors        []*Vendor
	UploadURL      *url.URL
}

func renderBillForm(ctx *Context, errs []string) error {
	uploadURL, err := blobstore.UploadURL(ctx.c, "/admin/bill/create", nil)
	if err != nil {
		return err
	}

	vendors, err := ctx.GetAllVendors()
	if err != nil {
		return err
	}
	return ctx.renderAdmin(newBillTmpl, NewBillForm{&Bill{}, errs, vendors, uploadURL})
}

func handleNewBill(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	return renderBillForm(ctx, []string{})
}

func handleCreateBill(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	errs := []string{}

	blobs, fields, err := blobstore.ParseUpload(ctx.r)

	ctx.Debugf("Blobs: %v", blobs)

	if err != nil {
		return err
	}

	//vendor := getFormFieldString(fields, "vendor")
	amt := getFormFieldInt(fields, "amount")
	if amt <= 0 {
		errs = append(errs, "Amount must be greater than 0")
	}

	vendorID := getFormFieldString(fields, "vendor")
	if vendorID == "" {
		errs = append(errs, "You must choose a vendor for the bill")
	}

	file := blobs["file"]
	if len(file) == 0 {
		errs = append(errs, "You must upload a bill file")
	}

	if len(errs) > 0 {
		return renderBillForm(ctx, errs)
	}

	v, err := ctx.GetVendorByID(vendorID)

	if err != nil {
		return err
	}

	b := Bill{
		Amt:        amt,
		PostedOn:   time.Now(),
		VendorKey:  v.Key,
		CompanyKey: v.CompanyKey,
		PostedBy:   ctx.user.String(),
		BlobKey:    file[0].BlobKey,
	}

	key := datastore.NewIncompleteKey(ctx.c, "Bill", v.CompanyKey)
	_, err = datastore.Put(ctx.c, key, &b)
	if err != nil {
		return err
	}
	ctx.Flash("New bill created successfully!")
	return ctx.Redirect("/admin/bills")
}

var (
	adminDashTmpl   = adminTmpl("dashboard.html")
	newCompanyTmpl  = adminTmpl("new_company.html")
	newUserTmpl     = adminTmpl("new_user.html")
	viewCompanyTmpl = adminTmpl("view_company.html")
	newVendorTmpl   = adminTmpl("new_vendor.html")
	viewVendorTmpl  = adminTmpl("view_vendor.html")
	viewCompanies   = adminTmpl("companies.html")
	viewUsers       = adminTmpl("users.html")
	viewVendors     = adminTmpl("vendors.html")
	viewBills       = adminTmpl("bills.html")
	newBillTmpl     = adminTmpl("new_bill.html")
)

func setupAdminRoutes(router *mux.Router) {
	router.Handle("/admin/dashboard", adminOnly(handleAdminDashboard))
	router.Handle("/admin/companies", adminOnly(handleAdminCompanies))
	router.Handle("/admin/users", adminOnly(handleAdminUsers))
	router.Handle("/admin/vendors", adminOnly(handleAdminVendors))
	router.Handle("/admin/bills", adminOnly(handleAdminBills))
	router.Handle("/admin/user/new", adminOnly(handleNewUser))
	router.Handle("/admin/user/create", adminOnly(handleCreateUser))
	router.Handle("/admin/user/delete", adminOnly(handleDeleteUser))
	router.Handle("/admin/company/new", adminOnly(handleNewCompany))
	router.Handle("/admin/company/create", adminOnly(handleCreateCompany))
	router.Handle("/admin/company/view", adminOnly(handleViewCompany))

	router.Handle("/admin/vendor/new", adminOnly(handleNewVendor))
	router.Handle("/admin/vendor/create", adminOnly(handleCreateVendor))
	router.Handle("/admin/vendor/view", adminOnly(handleViewVendor))

	router.Handle("/admin/bill/new", adminOnly(handleNewBill))
	router.Handle("/admin/bill/create", adminOnly(handleCreateBill))

}
