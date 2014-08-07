package billing

import (
	"appengine"
	"appengine/datastore"
	"time"
)

type Bill struct {
	ID         int64          `datastore:"-"`
	Key        *datastore.Key `datastore:"-"`
	PostedBy   string
	PostedOn   time.Time
	CompanyKey *datastore.Key
	VendorKey  *datastore.Key
	BlobKey    appengine.BlobKey
	Amt        int
	Paid       bool
	Reconciled bool

	Company *Company `datastore:"-"`
	Vendor  *Vendor  `datastore:"-"`
}

func (ctx *Context) GetAllBills() ([]*Bill, error) {
	var bills []*Bill
	q := datastore.NewQuery("Bill").Order("-PostedOn").Limit(10)
	bills = make([]*Bill, 0, 10)
	keys, err := q.GetAll(ctx.c, &bills)
	if err != nil {
		return bills, err
	}

	for idx, k := range keys {
		bills[idx].ID = k.IntID()
		bills[idx].Key = k
	}

	return bills, nil
}

func (ctx *Context) GetCompanyUnreconciledBills(c *Company) ([]*Bill, error) {

	var bills []*Bill
	q := datastore.NewQuery("Bill").Ancestor(c.Key).Filter("Reconciled = ", false).Order("-PostedOn").Limit(20)
	bills = make([]*Bill, 0, 20)
	keys, err := q.GetAll(ctx.c, bills)
	if err != nil {
		return bills, err
	}

	for idx, k := range keys {
		bills[idx].ID = k.IntID()
		bills[idx].Key = k
	}

	return bills, nil
}

func (ctx *Context) LoadBillCompanies(bills []*Bill) error {
	var keys []*datastore.Key
	for _, b := range bills {
		keys = append(keys, b.CompanyKey)
	}

	companies, err := ctx.GetCompanyMulti(keys)

	if err != nil {
		return err
	}

	for idx, b := range bills {
		b.Company = companies[idx]
	}

	return nil
}

func (ctx *Context) LoadBillVendors(bills []*Bill) error {
	var keys []*datastore.Key
	for _, b := range bills {
		keys = append(keys, b.VendorKey)
	}

	vendors, err := ctx.GetVendorMulti(keys)

	if err != nil {
		return err
	}

	for idx, b := range bills {
		b.Vendor = vendors[idx]
	}

	return nil
}

func (ctx *Context) GetCompanyReconciledBills(c *Company) ([]*Bill, error) {

	var bills []*Bill
	q := datastore.NewQuery("Bill").Ancestor(c.Key).Filter("Reconciled = ", true).Order("-PostedOn").Limit(20)
	bills = make([]*Bill, 0, 20)
	keys, err := q.GetAll(ctx.c, &bills)
	if err != nil {
		return bills, err
	}

	for idx, k := range keys {
		bills[idx].ID = k.IntID()
		bills[idx].Key = k
	}

	return bills, nil
}

func (ctx *Context) GetBillCount() (int, error) {
	c, err := datastore.NewQuery("Bill").Count(ctx.c)
	return c, err
}
