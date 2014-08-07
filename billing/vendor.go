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

func (ctx *Context) LoadVendorCompanies(vendors []*Vendor) error {
	var keys []*datastore.Key
	for _, v := range vendors {
		keys = append(keys, v.CompanyKey)
	}

	companies, err := ctx.GetCompanyMulti(keys)

	if err != nil {
		return err
	}

	for idx, v := range vendors {
		v.Company = companies[idx]
	}

	return nil
}

func (ctx *Context) GetVendorMulti(keys []*datastore.Key) ([]*Vendor, error) {
	vendors := make([]*Vendor, len(keys))

	for idx, _ := range vendors {
		vendors[idx] = new(Vendor)
	}

	err := datastore.GetMulti(ctx.c, keys, vendors)

	return vendors, err
}

func (ctx *Context) GetVendorByID(id string) (*Vendor, error) {
	v := new(Vendor)
	k, err := datastore.DecodeKey(id)

	v.Key = k

	if err != nil {
		return v, err
	}

	err = datastore.Get(ctx.c, k, v)

	return v, err
}

func (ctx *Context) GetVendorCount() (int, error) {
	c, err := datastore.NewQuery("Vendor").Count(ctx.c)
	return c, err
}
