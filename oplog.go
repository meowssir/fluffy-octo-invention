package oplog

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/net/websocket"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	origin := "http://localhost/"
	server := "ws://localhost:12345/ws"
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

// FIXME: unmarshal getter/setter.
func websocketDial(o Result) {
	config, err := websocket.NewConfig(server, origin)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: Evaluate handshake latency overhead with the additional 9 bytes of header data.
	config.Header.Add("Mongo", "true")
	ws, err := websocket.DialConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	ws.Write([]byte(o))
	if err != nil {
		log.Fatal(err)
	}
}

func tailOplog() {
	_ = "breakpoint"
	session, err := mgo.Dial(url)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	result := OplogEntry{}
	lastId := Result{}

	collection := session.DB("local").C("oplog.rs")
	iter := collection.Find(bson.M{"ns": "test.foo"}).Sort("$natural").Tail(5 * time.Second)
	for {
		for iter.Next(&result) {
			fmt.Println(result.O)
			lastId = result.O
			// TODO: connect and send an ObjectId. GetBSON() to unmarshal and decode to bytes.
			// FIXME: this should write not Dial every time we send an oplog entry.
			websocketDial(lastId)
		}
		if iter.Err() != nil {
			fmt.Println("an error happened")
		}
		if iter.Timeout() {
			continue
		}
		query := collection.Find(bson.M{"_id": bson.M{"$gt": lastId}})
		iter = query.Sort("$natural").Tail(5 * time.Second)
	}
	iter.Close()
}
