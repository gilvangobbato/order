package usecase

import (
	"context"
	"eda-example/internal/domain/entity"
	"eda-example/internal/domain/event"
	"eda-example/internal/domain/queue"
	"fmt"
	"time"
)

type ProcessOrderPaymentUseCase struct {
	publisher queue.Publisher
}

func NewProcessOrderPaymentUseCase(publisher queue.Publisher) *ProcessOrderPaymentUseCase {
	return &ProcessOrderPaymentUseCase{
		publisher: publisher,
	}
}

func (h *ProcessOrderPaymentUseCase) Execute(ctx context.Context, payload *event.OrderCreatedEvent) error {
	fmt.Println("--- ProcessOrderPaymentUseCase ---")

	// TODO: find order by id in the repository database here
	order, err := entity.RestoreOrderEntity(payload.Id, payload.Status)
	if err != nil {
		return err
	}

	for _, i := range payload.Items {
		item := entity.NewOrderItemEntity(i.ProductName, i.TotalPrice/float64(i.Quantity), i.Quantity)
		order.AddItem(item)
	}

	paymentValue := payload.TotalPrice
	err = order.Pay(paymentValue)
	if err != nil {
		return err
	}

	fmt.Printf("Order Paid. Value: %f \n", payload.TotalPrice)
	err = h.publisher.Publish(ctx, event.OrderPaidEvent{OrderId: payload.Id, PaidValue: paymentValue, PaymentDate: time.Now()})
	if err != nil {
		return err
	}
	return nil
}
