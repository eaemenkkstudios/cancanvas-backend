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
	AcceptBid(sender, auctionID, bidID string) (bool, error)
	AcceptedBids(sender string) ([]*model.Auction, error)
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
	for _, b := range auction.Bids {
		if b.Issuer == sender {
			return nil, errors.New("You can't have more than 1 bid in the same auction")
		}
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

func (db *auctionRepository) AcceptBid(sender, auctionID, bidID string) (bool, error) {
	id, err := primitive.ObjectIDFromHex(auctionID)
	if err != nil {
		return false, errors.New("Invalid auctionID")
	}
	collection := db.client.Collection(CollectionAuctions)
	result := collection.FindOne(context.TODO(), bson.M{"_id": id})
	var auction model.Auction
	err = result.Decode(&auction)
	if err != nil {
		return false, errors.New("Auction not found")
	}
	if auction.Host != sender {
		return false, errors.New("Unauthorized")
	}
	changed := false
	for i, b := range auction.Bids {
		if b.ID == bidID || !b.Selected {
			auction.Bids[i].Selected = true
			changed = true
			break
		}
	}
	if !changed {
		return true, nil
	}
	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.M{
		"$set": bson.M{"bids": auction.Bids},
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *auctionRepository) AcceptedBids(sender string) ([]*model.Auction, error) {
	collection := db.client.Collection(CollectionAuctions)
	ctx := context.TODO()
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	auctionList := make([]*model.Auction, 0)
	for cursor.Next(ctx) {
		var a model.Auction
		err = cursor.Decode(&a)
		if a.Host != sender {
			for _, b := range a.Bids {
				if b.Issuer == sender && b.Selected {
					auctionList = append(auctionList, &a)
				}
			}
		}
	}
	return auctionList, err
}

// NewAuctionRepository function
func NewAuctionRepository() AuctionRepository {
	client := newDatabaseClient()
	return &auctionRepository{
		client,
	}
}
