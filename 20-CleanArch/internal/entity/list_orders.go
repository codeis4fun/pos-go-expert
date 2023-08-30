package entity

import "errors"

var (
	errInvalidPage  = errors.New("invalid page")
	errInvalidLimit = errors.New("invalid limit")
)

type ListOrders struct {
	Page, Limit int
}

func NewListOrders(page, limit int) (*ListOrders, error) {
	listOrders := &ListOrders{
		Page:  page,
		Limit: limit,
	}
	err := listOrders.IsValid()
	if err != nil {
		return nil, err
	}
	return listOrders, nil
}

func (o *ListOrders) IsValid() error {
	if o.Page <= 0 {
		return errInvalidPage
	}
	if o.Limit <= 0 {
		return errInvalidLimit
	}
	return nil
}
