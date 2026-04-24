package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var clients = make(map[string]map[string]*websocket.Conn)
var mu sync.Mutex

func endConnection(gameID string, playerID string){
	mu.Lock()
	delete(clients[gameID], playerID)
	
 	isRoomEmpty := len(clients[gameID]) == 0
  	var remainingPlayers []*websocket.Conn;

   	for _, client := range clients[gameID]{
  		remainingPlayers = append(remainingPlayers, client)
    }
    
 	if isRoomEmpty { delete(clients, gameID) }
    mu.Unlock()
    
    if isRoomEmpty {
    	msg, err := deleteGame(gameID);
       	if err != nil{
    		log.Print("Trouble in deleting the game");
            return;
    	}
     	log.Print(msg);
    }else{ 
    	finalBoard, err := deletePlayer(gameID, playerID)
     	if err != nil{
      		log.Print("Trouble in deleting the player")
      	}
       	for _, client := range remainingPlayers{
        	client.WriteJSON(finalBoard);
        }
    }
}

func ws(w http.ResponseWriter, r *http.Request){
	//Connecting Websocket
	conn, err := upgrader.Upgrade(w, r, nil); 
	if err != nil{
		log.Print("Upgrade Error");
		return
	}
	defer conn.Close()
	
	
	//Getting Data
	gameID := r.PathValue("id")
	playerID := r.URL.Query().Get("player");
	if playerID == "" { playerID = "Anonymous"};
	req, err := gameChecker(gameID, playerID);
	if(err != nil){
		log.Print("Error: ", err);
		return
	}
    
    //Setting Data
    mu.Lock()
    if clients[gameID] == nil{
   		clients[gameID] = make(map[string]*websocket.Conn);
    }
    clients[gameID][playerID] = conn;
    mu.Unlock()
    
    defer endConnection(gameID, playerID);
    
    //Initial state
    err = conn.WriteJSON(&req)
    if err != nil {
    	log.Print("Error sending response to user");
     	return
    }
    
    for {
    	var move GameStruct
     	err = conn.ReadJSON(&move)
      	if err != nil{
     		if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway){
       			log.Print("Unexpected Error while getting data: ", err);
       		}
         	break
       	}
        latestBoard, err := getLatestBoard(gameID);
        if err != nil{
        	conn.WriteJSON(map[string]string{"error":"Could not sync game state"})
         	continue
        }
        latestBoard.Choice = move.Choice;
        nextMove, err := play(latestBoard, playerID);
        if err!=nil{
	       errorMessage := map[string]string{"error": err.Error()}
		   conn.WriteJSON(&errorMessage)
		   continue	
	    }
				
        fmt.Printf("Player %s made a move in game %s\n", playerID, gameID)
        
        mu.Lock()
        for _, client := range clients[gameID]{
        	client.WriteJSON(&nextMove);
        }
        mu.Unlock()
    }
}