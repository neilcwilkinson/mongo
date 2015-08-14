package mongo

import (
	"encoding/json"
	"fmt"

	"gopkg.in/mgo.v2"
)

var (
	ctl               *MongoSession
	db                *mgo.Database
	collection        *mgo.Collection
	mongouri          = ""
	connectionChannel = make(chan bool)
)

var (
	ConnectionChannel = make(chan bool)
	Connected         = false
)

type MongoSession struct {
	// This will be our extensible type that will
	// be used as a common context type for our routes
	session *mgo.Session // our cloneable session
}

func receiveConnectionStatus() {
	go func() {
		for {
			isConnected := <-connectionChannel
			fmt.Printf("%b\n", isConnected)
			if isConnected != Connected {
				fmt.Printf("State changed: %b\n", isConnected)
			}
			Connected = isConnected
			if Connected == false {
				//this only works in a go routine
				go Initialize(mongouri)
			}
			//send to any remote listeners
			ConnectionChannel <- isConnected
		}
	}()
}

func Initialize(uri string) error {
	go receiveConnectionStatus()

	mongouri = uri

	ctl, err := NewMongoSession(mongouri)

	if err != nil {
		connectionChannel <- false
		return err
	} else {
		db = ctl.session.Clone().DB("test")
		connectionChannel <- true
	}
	return nil
}

func NewMongoSession(uri string) (*MongoSession, error) {
	session, err := mgo.Dial(uri)
	if err != nil {
		connectionChannel <- false
		go Initialize(mongouri)
		return nil, err
	}
	connectionChannel <- true
	return &MongoSession{
		session: session,
	}, nil
}

//move this into caling procedure so as to be able to respect state changes?
func LogMessage(collectionName string, key string, message []byte) {
	if Connected == true {
		collection := db.C(collectionName)
		var m map[string]interface{}
		err := json.Unmarshal([]byte(message), &m)

		//info, err := collection.UpsertId(r.Product, r)
		info, err := collection.UpsertId(key, m)
		if err != nil {
			connectionChannel <- false
			fmt.Printf("Unable to upsert document:%v\n", err)
		} else {
			fmt.Sprintf("Upserted:", info.UpsertedId)
		}
	} else {
		fmt.Println("\nUnable to write to Mongo")
	}
}
