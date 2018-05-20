package main

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/net/websocket"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	server = "ws://localhost:12345/ws"
	origin = "http://localhost/"
	url    = "Byrons-MacBook-Pro-3.local:27017,Byrons-MacBook-Pro-3.local:27018"
)

type Result struct {
	Id bson.ObjectId `bson:"_id,omitempty"`
}

// This struct represents a BSON document containing the elements of an oplog entry.
type OplogEntry struct {
	Ts time.Time
	T  int64
	H  int64
	V  int64
	Op string
	Ns string
	O  Result
}

func websocketDial() *websocket.Conn {
	config, err := websocket.NewConfig(server, origin)
	if err != nil {
		log.Fatal(err)
	}
	// NOTE: Handshake latency overhead with the additional 9 bytes of header data is negligible.
	config.Header.Add("Mongo", "true")
	ws, err := websocket.DialConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	return ws
}

func main() {
	_ = "breakpoint"
	session, err := mgo.Dial(url)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var (
		result = OplogEntry{}
		lastId = Result{}
	)

	ws := websocketDial()
	
	collection := session.DB("local").C("oplog.rs")
	iter := collection.Find(bson.M{"ns": "test.foo"}).Sort("$natural").Tail(5 * time.Second)
	for {
		for iter.Next(&result) {
			fmt.Println(result.O)
			lastId = result.O
			a, err := lastId.Id.MarshalText()
			if err != nil {
				log.Fatal(err)
			}
			ws.Write(a)
		}
		if iter.Err() != nil {
			log.Println("error: %v", iter.Err())
		}
		if iter.Timeout() {
			continue
		}
		query := collection.Find(bson.M{"_id": bson.M{"$gt": lastId}})
		iter = query.Sort("$natural").Tail(5 * time.Second)
	}
	iter.Close()
}
