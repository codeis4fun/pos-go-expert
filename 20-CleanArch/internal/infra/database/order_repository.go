package database

import (
	"database/sql"

	"github.com/codeis4fun/pos-go-expert/20-CleanArch/internal/entity"
)

type OrderRepository struct {
	Db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{Db: db}
}

func (r *OrderRepository) Save(order *entity.Order) error {
	stmt, err := r.Db.Prepare("INSERT INTO orders (id, price, tax, final_price) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(order.ID, order.Price, order.Tax, order.FinalPrice)
	if err != nil {
		return err
	}
	return nil
}

func (r *OrderRepository) GetOrders(listOrders *entity.ListOrders) ([]entity.Order, error) {
	orders := []entity.Order{}
	offset := (listOrders.Page - 1) * listOrders.Limit
	rows, err := r.Db.Query("SELECT id, price, tax, final_price FROM orders LIMIT ? OFFSET ?", listOrders.Limit, offset)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var order entity.Order
		err := rows.Scan(&order.ID, &order.Price, &order.Tax, &order.FinalPrice)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}
