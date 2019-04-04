package db

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"in-share-server/config"
	"log"
	"os"
	"time"
)

const filesCollectionName string = "fs.files"
const chunksCollectionName string = "fs.chunks"

type DatabaseClient struct {
	mongo    *mongo.Client
	database *mongo.Database
	bucket   *gridfs.Bucket
}

func (m *DatabaseClient) Initialize(config config.Config) {

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_CONNECTION_STRING")))

	if err != nil {
		// error
		log.Fatal(err)
	}

	err = client.Ping(ctx, readpref.Primary())

	if err != nil {
		//
		log.Println(err)
		log.Fatal("Could not connect database")
	}

	log.Println("Connected to Database")
	m.mongo = client
	m.database = client.Database(config.DB.DBName)

	m.bucket, _ = gridfs.NewBucket(m.database, options.GridFSBucket())
}

func (m *DatabaseClient) Update(collectionName string, filter bson.M, object bson.M) error {
	object["updatedAt"] = time.Now()
	collection := m.database.Collection(collectionName)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err := collection.UpdateOne(ctx, filter, object)

	if err == nil {
		log.Println("Updated ")
	}
	return err
}

func (m *DatabaseClient) Insert(collectionName string, object bson.M) string {
	collection := m.database.Collection(collectionName)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	now := time.Now()

	object["createdAt"] = now
	object["updatedAt"] = now

	res, _ := collection.InsertOne(ctx, object)
	id := res.InsertedID

	fmt.Println("Inserted ", id)
	return fmt.Sprintln(id)
}

//TODO review upload process so as not load all the files in memory
func (m *DatabaseClient) InsertFile(file []byte, filename string) (interface{}, error) {

	fmt.Println("inserting file ", filename)

	stream, err := m.bucket.OpenUploadStream(filename)

	if err != nil {
		fmt.Println("an error", err)
		return primitive.ObjectID{}, err
	}

	_, err = stream.Write(file)

	fmt.Println("Close")

	stream.Close()

	return stream.FileID, err
}

func (m *DatabaseClient) DownloadFile(fileId primitive.ObjectID, chunkLength int, skip int64) ([]byte, error) {

	fmt.Println("download ", fileId, chunkLength)

	stream, err := m.bucket.OpenDownloadStream(fileId)

	if err != nil {
		fmt.Println("err", err)
		return nil, err
	}

	if skip != 0 {
		skipLength, err := stream.Skip(skip)
		fmt.Println("skip length", skipLength)

		if err != nil {
			return nil, err
		}
	}

	b := make([]byte, chunkLength)

	length, err := stream.Read(b)

	if err != nil {
		fmt.Println("stream error", err)
	}

	stream.Close()

	fmt.Println("length ", length)
	return b, nil
}

func (m *DatabaseClient) FindOne(collectionName string, filter bson.M, v interface{}) error {
	collection := m.database.Collection(collectionName)

	err := collection.FindOne(context.TODO(), filter).Decode(v)

	if err != nil {
		fmt.Println(err)
	}

	return err
}
