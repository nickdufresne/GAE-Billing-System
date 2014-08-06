package billing

import (
	"appengine"
	"appengine/datastore"
	"time"
)

type Bill struct {
	ID         string         `datastore:"-"`
	Key        *datastore.Key `datastore:"-"`
	PostedBy   string
	PostedOn   time.Time
	CompanyKey *datastore.Key
	VendorKey  *datastore.Key
	BlobKey    appengine.BlobKey
	Amt        int
	Paid       bool
	Reconciled bool
}

func (ctx *Context) GetCompanyUnreconciledBills(c *Company) ([]*Bill, error) {

	var bills []*Bill
	q := datastore.NewQuery("Bill").Ancestor(c.Key).Filter("Reconciled = ", false).Order("-PostedOn").Limit(20)
	bills = make([]*Bill, 0, 20)
	keys, err := q.GetAll(ctx.c, &bills)
	if err != nil {
		return bills, err
	}

	for idx, k := range keys {
		bills[idx].ID = k.Encode()
		bills[idx].Key = k
	}

	return bills, nil
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
		bills[idx].ID = k.Encode()
		bills[idx].Key = k
	}

	return bills, nil
}
