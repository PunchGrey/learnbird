package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Word struct {
	Eng       string    `bson:"eng"`
	Rus       string    `bson:"ru"`
	CreatedAt time.Time `bson:"created_at"`
}

func connection(url string, user string, password string) *mongo.Client {
	auth := options.Credential{
		Username: user,
		Password: password,
	}
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017").SetAuth(auth)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Println("error connection:", err)
		return nil
	}
	// Check the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Println("error check the connection:", err)
		return nil
	}
	log.Println("Connected to MongoDB!")
	return client

}

func Disconnect(client *mongo.Client) error {
	err := client.Disconnect(context.Background())
	if err != nil {
		return err
	}
	log.Println("disconnect from MongoDB!")
	return nil
}

func getData(client *mongo.Client, nameCollection string, depth int64) []Word {
	data := []Word{}

	// Select the database and collection
	collection := client.Database("vocabulary").Collection(nameCollection)

	filter := bson.M{}
	options := options.Find()
	options.SetSort(bson.M{"_id": -1})
	options.SetLimit(depth)

	// Find all documents in the collection
	cursor, err := collection.Find(context.Background(), filter, options)
	if err != nil {
		log.Println("Erorr find documents: ", err)
		return nil
	}

	defer cursor.Close(context.Background())

	// Iterate over the cursor and print each document
	for cursor.Next(context.Background()) {
		var raw bson.Raw
		err := cursor.Decode(&raw)
		if err != nil {
			log.Println(err)
			return nil
		}
		var ward Word

		err = bson.Unmarshal(raw, &ward)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		data = append(data, ward)
	}
	return data
}
