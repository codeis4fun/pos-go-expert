package usecase

import (
	"github.com/devfullcycle/20-CleanArch/internal/entity"
)

type ListOrdersInputDTO struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

type ListOrdersOutputDTO struct {
	Orders []OrderOutputDTO `json:"orders"`
}

type ListOrdersUseCase struct {
	OrderRepository entity.OrderRepositoryInterface
}

func NewListOrdersUseCase(OrderRepository entity.OrderRepositoryInterface) *ListOrdersUseCase {
	return &ListOrdersUseCase{
		OrderRepository: OrderRepository,
	}
}

func (lo *ListOrdersUseCase) Execute(input ListOrdersInputDTO) (ListOrdersOutputDTO, error) {
	listOrders, err := entity.NewListOrders(input.Page, input.Limit)
	if err != nil {
		return ListOrdersOutputDTO{}, err
	}

	orders, err := lo.OrderRepository.GetOrders(listOrders)
	if err != nil {
		return ListOrdersOutputDTO{}, err
	}

	dto := ListOrdersOutputDTO{}
	for _, order := range orders {
		dto.Orders = append(dto.Orders, OrderOutputDTO{
			ID:         order.ID,
			Price:      order.Price,
			Tax:        order.Tax,
			FinalPrice: order.FinalPrice,
		})
	}

	return dto, nil
}
