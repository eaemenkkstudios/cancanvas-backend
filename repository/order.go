package repository

import (
	"context"
	"errors"
	"strconv"

	"github.com/eaemenkkstudios/cancanvas-backend/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// OrderRepository interface
type OrderRepository interface {
	CreateOrder(auctionID, bidID, description string, price float64) (string, error)
	UpdateOrder(orderID, status string, payerID *string) (bool, error)
	DeleteOrder(orderID string) (bool, error)
}

type orderRepository struct {
	client *mongo.Database
}

// Payment struct
type Payment struct {
	ID         string  `bson:"_id,omitempty"`
	PaymentID  string  `bson:"paymentID"`
	PaymentURL string  `bson:"paymentURL"`
	PayerID    *string `bson:"payerID"`
	AuctionID  string  `bson:"auctionID"`
	BidID      string  `bson:"bidID"`
	Status     string  `bson:"status"`
}

func (db *orderRepository) CreateOrder(auctionID, bidID, description string, price float64) (string, error) {
	collection := db.client.Collection(CollectionPayments)
	result := collection.FindOne(context.TODO(), bson.M{"auctionID": auctionID, "bidID": bidID})
	var p Payment
	err := result.Decode(&p)
	if err == nil {
		switch p.Status {
		case "PENDING":
			return p.PaymentURL, nil
		case "COMPLETED":
			return "", errors.New("Could not generate link")
		}
	}
	paypalClient := service.NewPayPalService()
	order, err := paypalClient.CreateOrder(strconv.FormatFloat(price, 'f', 2, 64), description)
	if err != nil {
		return "", err
	}
	_, err = collection.InsertOne(context.TODO(), Payment{
		PaymentID:  order.ID,
		PaymentURL: order.URL,
		AuctionID:  auctionID,
		BidID:      bidID,
		Status:     "PENDING",
	})
	if err != nil {
		return "", err
	}
	return order.URL, nil
}

func (db *orderRepository) UpdateOrder(orderID, status string, payerID *string) (bool, error) {
	collection := db.client.Collection(CollectionPayments)
	result, err := collection.UpdateOne(context.TODO(), bson.M{"paymentID": orderID}, bson.M{
		"$set": bson.M{"status": status, "payerID": *payerID},
	})
	if err != nil || result.ModifiedCount == 0 {
		return false, errors.New("Order not found")
	}
	return true, nil
}

func (db *orderRepository) DeleteOrder(orderID string) (bool, error) {
	collection := db.client.Collection(CollectionPayments)
	result := collection.FindOne(context.TODO(), bson.M{"paymentID": orderID})
	var p Payment
	err := result.Decode(&p)
	if err != nil {
		return false, errors.New("Order not found")
	}
	if p.Status != "COMPLETED" {
		result, err := collection.DeleteOne(context.TODO(), p)
		if err != nil && result.DeletedCount > 0 {
			return true, nil
		}
		return false, err
	}
	return false, errors.New("Cannot delete a completed order")
}

// NewOrderRepository function
func NewOrderRepository() OrderRepository {
	client := newDatabaseClient()
	return &orderRepository{
		client,
	}
}
