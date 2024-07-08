package main

import (
	"context"
	"eda-example/internal/application/controller"
	"eda-example/internal/application/usecase"
	"eda-example/internal/domain/event"
	"eda-example/internal/infra/queue"
	"fmt"
	"log"
	"net/http"
	"reflect"
)

func main() {
	ctx := context.Background()

	// initialize queue
	queue := queue.NewRabbitMQAdapter("amqp://guest:guest@localhost:5672/")

	// use cases
	createOrderUseCase := usecase.NewCreateOrderUseCase(queue)
	processPaymentUseCase := usecase.NewProcessOrderPaymentUseCase(queue)
	stockMovementUseCase := usecase.NewStockMovementUseCase()
	sendOrderEmailUseCase := usecase.NewSendOrderEmailUseCase()

	// controllers
	orderController := controller.NewOrderController(createOrderUseCase, processPaymentUseCase, stockMovementUseCase, sendOrderEmailUseCase)

	// register routes
	http.HandleFunc("POST /create-order", orderController.CreateOrder)

	// mapping listeners
	var list map[reflect.Type][]func(w http.ResponseWriter, r *http.Request) = map[reflect.Type][]func(w http.ResponseWriter, r *http.Request){
		reflect.TypeOf(event.OrderCreatedEvent{}): {
			orderController.ProcessOrderPayment,
			orderController.StockMovement,
			orderController.SendOrderEmail,
		},
	}

	// register listeners
	for eventType, handlers := range list {
		for _, handler := range handlers {
			queue.ListenerRegister(eventType, handler)
		}
	}

	// connect queue
	err := queue.Connect(ctx)
	if err != nil {
		log.Fatalf("Error connect queue %s", err)
	}
	defer queue.Disconnect(ctx)

	// start consuming queues
	OrderCreatedEvent := reflect.TypeOf(event.OrderCreatedEvent{}).Name()

	go func(ctx context.Context, queueName string) {
		err = queue.StartConsuming(ctx, queueName)
		if err != nil {
			log.Fatalf("Error running consumer %s: %s", queueName, err)
		}
	}(ctx, OrderCreatedEvent)

	// start server
	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}
