package billing

import (
	"appengine/datastore"
	"time"
)

type Vendor struct {
	ID         string         `datastore:"-"`
	Key        *datastore.Key `datastore:"-"`
	CompanyKey *datastore.Key
	Name       string
	CreatedOn  time.Time
	CreatedBy  string
	Company    *Company `datastore:"-"`
}

func (ctx *Context) GetAllVendors() ([]*Vendor, error) {
	var vendors []*Vendor
	q := datastore.NewQuery("Vendor").Order("Name").Limit(10)
	vendors = make([]*Vendor, 0, 10)
	keys, err := q.GetAll(ctx.c, &vendors)
	if err != nil {
		return vendors, err
	}

	for idx, k := range keys {
		vendors[idx].ID = k.Encode()
		vendors[idx].Key = k
	}

	return vendors, nil
}

func (ctx *Context) GetCompanyVendors(c *Company) ([]*Vendor, error) {

	var vendors []*Vendor
	q := datastore.NewQuery("Vendor").Ancestor(c.Key).Order("Name").Limit(20)
	vendors = make([]*Vendor, 0, 20)
	keys, err := q.GetAll(ctx.c, &vendors)
	if err != nil {
		return vendors, err
	}

	for idx, k := range keys {
		vendors[idx].ID = k.Encode()
		vendors[idx].Key = k
	}

	return vendors, nil
}
