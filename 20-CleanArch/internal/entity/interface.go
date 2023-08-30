package entity

type OrderRepositoryInterface interface {
	Save(order *Order) error
	GetOrders(listOrders *ListOrders) ([]Order, error)
}
