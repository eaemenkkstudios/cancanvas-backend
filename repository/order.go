package repository

import (
	"context"
	"errors"
	"strconv"

	"github.com/eaemenkkstudios/cancanvas-backend/graph/model"
	"github.com/eaemenkkstudios/cancanvas-backend/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// OrderRepository interface
type OrderRepository interface {
	CreateOrder(auctionID, bidID, description string, price float64) (string, error)
	GetOrder(sender, orderID string) (*model.Order, error)
	GetOrders(sender string) ([]*model.Order, error)
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

func (db *orderRepository) GetOrder(sender, orderID string) (*model.Order, error) {
	collection := db.client.Collection(CollectionPayments)
	result := collection.FindOne(context.TODO(), bson.M{"paymentID": orderID})
	var p Payment
	err := result.Decode(&p)
	if err != nil {
		return nil, errors.New("Order not found")
	}
	collection = db.client.Collection(CollectionAuctions)
	id, err := primitive.ObjectIDFromHex(p.AuctionID)
	if err != nil {
		return nil, errors.New("Error while getting order")
	}
	result = collection.FindOne(context.TODO(), bson.M{"_id": id})
	var a model.Auction
	err = result.Decode(&a)
	if err != nil {
		return nil, errors.New("Error while getting order")
	}
	order := &model.Order{
		ID:         p.ID,
		AuctionID:  p.AuctionID,
		BidID:      p.BidID,
		PayerID:    p.PayerID,
		PaymentID:  p.PaymentID,
		PaymentURL: p.PaymentURL,
		Status:     p.Status,
	}
	if a.Host == sender {
		return order, nil
	}
	for _, b := range a.Bids {
		if b.ID == p.BidID && b.Issuer == sender {
			return order, nil
		}
	}
	return nil, errors.New("Unauthorized")
}

func (db *orderRepository) GetOrders(sender string) ([]*model.Order, error) {
	collection := db.client.Collection(CollectionAuctions)
	ctx := context.TODO()
	cursor, err := collection.Find(ctx, bson.M{"host": sender})
	if err != nil {
		return nil, errors.New("No orders found")
	}
	auctionIDs := make([]string, 0)
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var a model.Auction
		err = cursor.Decode(&a)
		if err != nil {
			continue
		}
		auctionIDs = append(auctionIDs, a.ID)
	}

	collection = db.client.Collection(CollectionPayments)
	cursor, err = collection.Find(ctx, bson.M{"auctionID": bson.M{"$in": auctionIDs}})
	orders := make([]*model.Order, 0)
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var o model.Order
		err = cursor.Decode(&o)
		orders = append(orders, &o)
	}
	return orders, err
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
