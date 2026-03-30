package repo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type IMongoDatabase interface {
	Collection(name string) IMongoCollection
}

type IMongoCollection interface {
	Find(ctx context.Context, filter interface{}, opts ...options.Lister[options.FindOptions]) (IMongoCursor, error)
	FindOne(ctx context.Context, filter interface{}, opts ...options.Lister[options.FindOneOptions]) IMongoSingleResult
	UpdateByID(ctx context.Context, id any, update any, opts ...options.Lister[options.UpdateOneOptions]) (IMongoUpdateResult, error)
	FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...options.Lister[options.FindOneAndUpdateOptions]) IMongoSingleResult
	InsertOne(ctx context.Context, document any, opts ...options.Lister[options.InsertOneOptions]) (IMongoInsertOneResult, error)
	DeleteOne(ctx context.Context, filter any, opts ...options.Lister[options.DeleteOneOptions]) (IDeleteResult, error)
}

type IMongoSingleResult interface {
	Decode(v interface{}) error
	Err() error
}

type IMongoUpdateResult interface {}

type IMongoInsertOneResult interface {
	InsertedId() any
}

type IDeleteResult interface {}

type IMongoCursor interface {
	Close(context.Context) error
	Next(context.Context) bool
	Decode(interface{}) error
	All(context.Context, interface{}) error
}

type MongoCollection struct {
	Сoll *mongo.Collection
}

type MongoSingleResult struct {
	sr *mongo.SingleResult
}

type MongoUpdateResult struct {
	ur *mongo.UpdateResult
}

type MongoInsertOneResult struct {
	ir         *mongo.InsertOneResult
	InsertedID any
}

type MongoDeleteOneResult struct {
	dr *mongo.DeleteResult
}

type MongoCursor struct {
	cur *mongo.Cursor
}

func (msr *MongoSingleResult) Decode(v interface{}) error {
	return msr.sr.Decode(v)
}
func (msr *MongoSingleResult) Err() error {
	return msr.sr.Err()
}

func (mir *MongoInsertOneResult) InsertedId() any {
	return mir.InsertedID
}

func (mc *MongoCursor) Close(ctx context.Context) error {
	return mc.cur.Close(ctx)
}

func (mc *MongoCursor) Next(ctx context.Context) bool {
	return mc.cur.Next(ctx)
}

func (mc *MongoCursor) Decode(val interface{}) error {
	return mc.cur.Decode(val)
}

func (mc *MongoCursor) All(ctx context.Context, results any) error {
	return mc.cur.All(ctx, results)
}

func (mc *MongoCursor) Err() error {
	return mc.cur.Err()
}

func (mc *MongoCollection) Find(ctx context.Context, filter interface{}, opts ...options.Lister[options.FindOptions]) (IMongoCursor, error) {
	cursorResult, err := mc.Сoll.Find(ctx, filter, opts...)
	return &MongoCursor{cur: cursorResult}, err
}

func (mc *MongoCollection) FindOne(ctx context.Context, filter interface{}, opts ...options.Lister[options.FindOneOptions]) IMongoSingleResult {
	singleResult := mc.Сoll.FindOne(ctx, filter, opts...)
	return &MongoSingleResult{sr: singleResult}
}

func (mc *MongoCollection) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...options.Lister[options.FindOneAndUpdateOptions]) IMongoSingleResult {
	singleResult := mc.Сoll.FindOneAndUpdate(ctx, filter, update, opts...)
	return &MongoSingleResult{sr: singleResult}
}

func (mc *MongoCollection) UpdateByID(ctx context.Context, id any, update any, opts ...options.Lister[options.UpdateOneOptions]) (IMongoUpdateResult, error) {
	updateResult, err := mc.Сoll.UpdateByID(ctx, id, update, opts...)
	return &MongoUpdateResult{ur: updateResult}, err
}

func (mc *MongoCollection) InsertOne(ctx context.Context, document any, opts ...options.Lister[options.InsertOneOptions]) (IMongoInsertOneResult, error) {
	insertRes, err := mc.Сoll.InsertOne(ctx, document, opts...)
	return &MongoInsertOneResult{ir: insertRes, InsertedID: insertRes.InsertedID}, err
}

func (mc *MongoCollection) DeleteOne(ctx context.Context, filter any, opts ...options.Lister[options.DeleteOneOptions]) (IDeleteResult, error) {
	delRes, err := mc.Сoll.DeleteOne(ctx, filter, opts...)
	return &MongoDeleteOneResult{dr: delRes}, err
}
