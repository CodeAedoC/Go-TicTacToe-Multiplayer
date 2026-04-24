package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"context"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var gameCollection *mongo.Collection;

type GameStruct struct {
	GameID     string		 `json:"gameid" bson:"_id,omitempty"`
	PlayerX    string		 `json:"playerx" bson:"playerx"`
	PlayerO    string		 `json:"playero" bson:"playero"`
	Board      [3][3]string  `json:"board" bson:"board"`
	Turn       string        `json:"turn" bson:"turn"`
	Win        bool          `json:"win" bson:"win"`
	Tie 	   bool 		 `json:"tie" bson:"tie"`
	TurnNumber int           `json:"turnNumber" bson:"turnNumber"`
	Choice     int           `json:"choice" bson:"-"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func (r *http.Request) bool{
		return true;
	},
}

func main() {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1);
	bsonOpts := &options.BSONOptions{
        ObjectIDAsHexString: true,
    }
    err := godotenv.Load()
    if(err != nil){
   		panic("Unable to get env file");
    }
	opts := options.Client().ApplyURI(os.Getenv("MONGO_URI")).SetServerAPIOptions(serverAPI).SetBSONOptions(bsonOpts);
	client, err := mongo.Connect(opts);
	if(err != nil){
		panic(err)
	}
	
	defer client.Disconnect(context.TODO());
	
	err = client.Ping(context.TODO(), nil);
	if(err != nil){
		panic(err)
	}
	gameCollection = client.Database("Tictactoe").Collection("Games");
	fmt.Println("Mongo ka bhosda Aaaag")
	
	http.HandleFunc("/start", start)
	// http.HandleFunc("/reset", reset)
	http.HandleFunc("/ws/{id}", ws)

	log.Fatal(http.ListenAndServe(":8080", nil))
}