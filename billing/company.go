package billing

import (
	"appengine"
	"appengine/datastore"

	"time"
)

type Company struct {
	ID        string         `datastore:"-"`
	Key       *datastore.Key `datastore:"-"`
	Name      string
	CreatedBy string
	CreatedOn time.Time
	Users     []*User
	Vendors   []*Vendor
}

// defaultCompanyKey returns the key used for all company entries.
func defaultCompanyKey(c appengine.Context) *datastore.Key {
	// The string "all_companies" here could be varied to have multiple base companies.
	return datastore.NewKey(c, "Company", "all_companies", 0, nil)
}

func (ctx *Context) GetAllCompanies() ([]*Company, error) {
	var companies []*Company
	q := datastore.NewQuery("Company").Ancestor(defaultCompanyKey(ctx.c)).Order("Name").Limit(10)
	companies = make([]*Company, 0, 10)
	keys, err := q.GetAll(ctx.c, &companies)
	if err != nil {
		return companies, err
	}

	for idx, k := range keys {
		companies[idx].ID = k.Encode()
		companies[idx].Key = k
	}

	return companies, nil
}

func (ctx *Context) GetCompanyByID(id string) (*Company, error) {
	c := new(Company)
	k, err := datastore.DecodeKey(id)

	c.Key = k

	if err != nil {
		return c, err
	}

	err = datastore.Get(ctx.c, k, c)

	return c, err
}
