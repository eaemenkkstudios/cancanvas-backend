package repository

import (
	"context"
	"errors"
	"time"

	"github.com/eaemenkkstudios/cancanvas-backend/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AuctionRepository interface
type AuctionRepository interface {
	GetAuctions(page *int) ([]*model.Auction, error)
	CreateAuction(sender, description string, offer float64) (*model.Auction, error)
	CreateBid(sender, auctionID, deadline string, price float64) (*model.Bid, error)
}

type auctionRepository struct {
	client *mongo.Database
}

func (db *auctionRepository) CreateAuction(sender, description string, offer float64) (*model.Auction, error) {
	auction := &model.Auction{
		Host:        sender,
		Offer:       offer,
		Description: description,
		Bids:        make([]*model.Bid, 0),
		Deadline:    time.Now().Add(72 * time.Hour),
		Timestamp:   time.Now(),
	}
	collection := db.client.Collection(CollectionAuctions)
	result, err := collection.InsertOne(context.TODO(), auction)
	if err != nil {
		return nil, err
	}
	id := result.InsertedID.(primitive.ObjectID).Hex()
	auction.ID = id
	return auction, nil
}

func (db *auctionRepository) GetAuctions(page *int) ([]*model.Auction, error) {
	if page == nil || *page < 1 {
		*page = 1
	}
	collection := db.client.Collection(CollectionAuctions)
	opts := options.Find().
		SetSkip(int64(PageSize * (*page - 1))).
		SetLimit(PageSize).
		SetSort(bson.M{"timestamp": -1})
	ctx := context.TODO()
	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, errors.New("Could not load posts")
	}
	var auctions = make([]*model.Auction, 0)
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var a model.Auction
		err = cursor.Decode(&a)
		auctions = append(auctions, &a)
	}
	return auctions, err
}

func (db *auctionRepository) CreateBid(sender, auctionID, deadline string, price float64) (*model.Bid, error) {
	collection := db.client.Collection(CollectionAuctions)
	id, err := primitive.ObjectIDFromHex(auctionID)
	if err != nil {
		return nil, errors.New("Invalid auctionID")
	}
	result := collection.FindOne(context.TODO(), bson.M{"_id": id})
	var auction model.Auction
	err = result.Decode(&auction)
	if err != nil {
		return nil, errors.New("Auction not found")
	}
	if auction.Host == sender {
		return nil, errors.New("You can't make a bid in your own auction")
	}
	if auction.Deadline.Unix() < time.Now().Unix() {
		return nil, errors.New("This auction is no longer accepting Bids")
	}
	bid := &model.Bid{
		ID:        primitive.NewObjectID().Hex(),
		Issuer:    sender,
		Deadline:  deadline,
		Price:     price,
		Timestamp: time.Now(),
	}
	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.M{
		"$push": bson.M{"Bids": bid},
	})
	if err != nil {
		return nil, errors.New("Unexpected error")
	}
	return bid, nil
}

// NewAuctionRepository function
func NewAuctionRepository() AuctionRepository {
	client := newDatabaseClient()
	return &auctionRepository{
		client,
	}
}
