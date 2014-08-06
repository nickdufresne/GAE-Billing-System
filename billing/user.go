package billing

import (
	"appengine/datastore"
	"time"
)

type User struct {
	ID          string         `datastore:"-"`
	Key         *datastore.Key `datastore:"-"`
	Email       string
	CompanyKey  *datastore.Key
	LastLoginOn time.Time
	CreatedOn   time.Time
	CreatedBy   string
	Company     *Company `datastore:"-"`
}

func (ctx *Context) GetAllUsers() ([]*User, error) {
	var users []*User
	q := datastore.NewQuery("User").Order("-LastLoginOn").Limit(10)
	users = make([]*User, 0, 10)
	keys, err := q.GetAll(ctx.c, &users)
	if err != nil {
		return users, err
	}

	for idx, k := range keys {
		users[idx].ID = k.Encode()
		users[idx].Key = k
	}

	return users, nil
}

func (ctx *Context) GetCompanyUsers(c *Company) ([]*User, error) {

	var users []*User
	q := datastore.NewQuery("User").Ancestor(c.Key).Order("Email").Limit(20)
	users = make([]*User, 0, 20)
	keys, err := q.GetAll(ctx.c, &users)
	if err != nil {
		return users, err
	}

	for idx, k := range keys {
		users[idx].ID = k.Encode()
		users[idx].Key = k
	}

	return users, nil
}

func (ctx *Context) UserEmailExists(e string) (bool, error) {
	q := datastore.NewQuery("User").Filter("Email =", e)
	cnt, err := q.Count(ctx.c)

	if err != nil {
		return false, err
	}

	if cnt > 0 {
		return true, nil
	}

	return false, nil
}

func (ctx *Context) GetUserAndCompanyByEmail(e string) (*User, *Company, error) {
	users := make([]*User, 0, 1)
	q := datastore.NewQuery("User").Filter("Email =", e).Limit(1)
	keys, err := q.GetAll(ctx.c, &users)
	if err != nil || len(keys) == 0 {
		return nil, nil, err
	}

	company := new(Company)
	err = datastore.Get(ctx.c, users[0].CompanyKey, company)
	if err != nil {
		return nil, nil, err
	}

	return users[0], company, err
}

func (ctx *Context) GetUserByID(id string) (*User, error) {
	user := new(User)
	k, err := datastore.DecodeKey(id)

	user.Key = k

	if err != nil {
		return user, err
	}

	err = datastore.Get(ctx.c, k, user)

	return user, err
}

func (ctx *Context) DeleteUser(id string) error {
	k, err := datastore.DecodeKey(id)

	if err != nil {
		return err
	}

	err = datastore.Delete(ctx.c, k)

	return err
}
