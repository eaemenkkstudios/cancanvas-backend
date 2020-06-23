package repository

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/eaemenkkstudios/cancanvas-backend/graph/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// AuctionRepository interface
type AuctionRepository interface {
	GetAuctions(page *int) ([]*model.FeedAuction, error)
	CreateAuction(sender, description string, offer float64) (*model.Auction, error)
	DeleteAuction(sender, auctionID string) (bool, error)
	CreateBid(sender, auctionID, deadline string, price float64) (*model.Bid, error)
	DeleteBid(sender, auctionID, bidID string) (bool, error)
	AcceptBid(sender, auctionID, bidID string) (bool, error)
	AcceptedBids(sender string) ([]*model.FeedAuction, error)
	BidPaymentLink(sender, auctionID, bidID string) (string, error)
}

type auctionRepository struct {
	client *mongo.Database
	order  OrderRepository
}

type feedAuction struct {
	ID          string        `bson:"_id"`
	Host        []*UserSchema `bson:"host"`
	Description string        `bson:"description"`
	Offer       float64       `bson:"offer"`
	Bids        []*model.Bid  `bson:"bids"`
	Timestamp   string        `bson:"timestamp"`
	Deadline    string        `bson:"deadline"`
}

func (db *auctionRepository) CreateAuction(sender, description string, offer float64) (*model.Auction, error) {
	auction := &model.Auction{
		Host:        sender,
		Offer:       offer,
		Description: description,
		Bids:        make([]*model.Bid, 0),
		Deadline:    strconv.FormatInt(time.Now().Add(72*time.Hour).Unix(), 10),
		Timestamp:   strconv.FormatInt(time.Now().Unix(), 10),
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

func (db *auctionRepository) DeleteAuction(sender, auctionID string) (bool, error) {
	collection := db.client.Collection(CollectionAuctions)
	id, err := primitive.ObjectIDFromHex(auctionID)
	if err != nil {
		return false, errors.New("Invalid auctionID")
	}
	result, err := collection.DeleteOne(context.TODO(), bson.M{"_id": id, "host": sender})
	if err != nil {
		return false, err
	}
	if result.DeletedCount == 0 {
		return false, errors.New("Could not delete auction")
	}
	return true, nil
}

func (db *auctionRepository) GetAuctions(page *int) ([]*model.FeedAuction, error) {
	if page == nil || *page < 1 {
		*page = 1
	}
	collection := db.client.Collection(CollectionAuctions)
	ctx := context.TODO()
	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{
		bson.D{{
			Key: "$lookup",
			Value: bson.M{
				"from":         CollectionUsers,
				"localField":   "host",
				"foreignField": "_id",
				"as":           "host",
			}}},
		bson.D{{Key: "$skip", Value: int64(PageSize * (*page - 1))}},
		bson.D{{Key: "$limit", Value: PageSize}},
		bson.D{{Key: "$sort", Value: bson.M{"timestamp": -1}}},
	})
	if err != nil {
		return nil, errors.New("Could not load posts")
	}
	var auctions = make([]*model.FeedAuction, 0)
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var a feedAuction
		err = cursor.Decode(&a)
		auctions = append(auctions, &model.FeedAuction{
			ID: a.ID,
			Host: &model.FeedUser{
				Name:     a.Host[0].Name,
				Nickname: a.Host[0].Nickname,
				Picture:  a.Host[0].Picture,
			},
			Bids:        a.Bids,
			Deadline:    a.Deadline,
			Description: a.Description,
			Offer:       a.Offer,
			Timestamp:   a.Timestamp,
		})
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
	auctionDeadline, err := strconv.ParseInt(auction.Deadline, 10, 64)
	if err != nil {
		return nil, err
	}
	if auctionDeadline < time.Now().Unix() {
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
		Timestamp: strconv.FormatInt(time.Now().Unix(), 10),
	}
	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.M{
		"$push": bson.M{"bids": bid},
	})
	if err != nil {
		return nil, errors.New("Unexpected error")
	}
	return bid, nil
}

func (db *auctionRepository) DeleteBid(sender, auctionID, bidID string) (bool, error) {
	collection := db.client.Collection(CollectionAuctions)
	id, err := primitive.ObjectIDFromHex(auctionID)
	if err != nil {
		return false, errors.New("Invalid auctionID")
	}
	result := collection.FindOne(context.TODO(), bson.M{"_id": id})
	var auction model.Auction
	err = result.Decode(&auction)
	if err != nil {
		return false, err
	}
	deleted := false
	for i, b := range auction.Bids {
		if b.ID == bidID && b.Issuer == sender && !b.Selected {
			auction.Bids = append(auction.Bids[:i], auction.Bids[i+1:]...)
			deleted = true
			break
		}
	}
	if !deleted {
		return false, errors.New("Could not delete bid")
	}
	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.M{
		"$set": bson.M{"bids": auction.Bids},
	})
	if err != nil {
		return false, errors.New("Could not delete bid")
	}
	return true, nil
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

func (db *auctionRepository) AcceptedBids(sender string) ([]*model.FeedAuction, error) {
	collection := db.client.Collection(CollectionAuctions)
	ctx := context.TODO()
	cursor, err := collection.Find(ctx, mongo.Pipeline{
		bson.D{{
			Key: "$lookup",
			Value: bson.M{
				"from":         CollectionUsers,
				"localField":   "host",
				"foreignField": "_id",
				"as":           "host",
			}}},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	auctionList := make([]*model.FeedAuction, 0)
	for cursor.Next(ctx) {
		var a feedAuction
		err = cursor.Decode(&a)
		if a.Host[0].Nickname != sender {
			for _, b := range a.Bids {
				if b.Issuer == sender && b.Selected {
					auctionList = append(auctionList, &model.FeedAuction{
						ID: a.ID,
						Host: &model.FeedUser{
							Name:     a.Host[0].Name,
							Nickname: a.Host[0].Nickname,
							Picture:  a.Host[0].Picture,
						},
						Bids:        a.Bids,
						Deadline:    a.Deadline,
						Description: a.Description,
						Offer:       a.Offer,
						Timestamp:   a.Timestamp,
					})
				}
			}
		}
	}
	return auctionList, err
}

func (db *auctionRepository) BidPaymentLink(sender, auctionID, bidID string) (string, error) {
	id, err := primitive.ObjectIDFromHex(auctionID)
	if err != nil {
		return "", errors.New("Invalid auctionID")
	}
	collection := db.client.Collection(CollectionAuctions)
	result := collection.FindOne(context.TODO(), bson.M{"_id": id})
	var auction model.Auction
	err = result.Decode(&auction)
	if err != nil {
		return "", errors.New("Auction not found")
	}
	if auction.Host != sender {
		return "", errors.New("Unauthorized")
	}

	for _, b := range auction.Bids {
		if b.ID == bidID {
			if !b.Selected {
				return "", errors.New("You need to accept this bid first")
			}
			url, err := db.order.CreateOrder(auctionID, bidID, auction.Description, b.Price)
			if err != nil {
				return "", err
			}
			return url, nil
		}
	}
	return "", errors.New("Could not generate link")
}

// NewAuctionRepository function
func NewAuctionRepository() AuctionRepository {
	client := newDatabaseClient()
	order := NewOrderRepository()
	return &auctionRepository{
		client,
		order,
	}
}
