package mongo

import (
	"encoding/json"
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	ctl               *MongoSession
	db                *mgo.Database
	collection        *mgo.Collection
	mongouri          = ""
	connectionChannel = make(chan bool)
)

//Exposed for calling go code by capitalizing the first character.
var (
	Connected = false
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

			if isConnected != Connected {
				fmt.Printf("State changed: %b\n", isConnected)
			}
			Connected = isConnected

			if Connected == false {
				//this only works in a go routine
				go Initialize(mongouri)
			}
		}
	}()
}

func Initialize(uri string) {
	mongouri = uri

	go receiveConnectionStatus()

	ctl, err := NewMongoSession(mongouri)
	if err != nil {
		connectionChannel <- false
	} else {
		db = ctl.session.Clone().DB("test")
		connectionChannel <- true
	}
	fmt.Println("\nReset to false")
}

func NewMongoSession(uri string) (*MongoSession, error) {
	// fmt.Printf("\nWTF\n")
	session, err := mgo.Dial(uri)
	if err != nil {
		return nil, err
	}
	return &MongoSession{
		session: session,
	}, nil
}

//move this into caling procedure so as to be able to respect state changes?
func LogMessage(collectionName string, key string, message []byte) {
	if Connected == true {

		var m map[string]interface{}
		err := json.Unmarshal([]byte(message), &m)
		m["createdAt"] = bson.Now()

		//info, err := collection.UpsertId(r.Product, r)
		collection := db.C(collectionName)
		info, err := collection.UpsertId(key, m)
		if err != nil {
			fmt.Printf("Unable to upsert document:%v\n", err)
			db.Session.Refresh()
			connectionChannel <- false
		} else {
			fmt.Sprintf("Upserted:", info.UpsertedId)
		}

		rates_log_collection := db.C("rateslog")
		err = rates_log_collection.Insert(m)
		if err != nil {
			fmt.Printf("Unable to insert document:%v\n", err)
			db.Session.Refresh()
			connectionChannel <- false
		}
		// else {
		// 	fmt.Sprintf("Upserted:", info.UpsertedId)
		// }

	} else {
		// fmt.Println("No connection")
	}
}
