package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func start(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	req := GameStruct{
		Turn: "X",
	}
	result, err := gameCollection.InsertOne(context.TODO(), req)
	if err != nil {
		http.Error(w, "Could not make game", 500)
		return
	}
	if oid, ok := result.InsertedID.(bson.ObjectID); ok {
		req.GameID = oid.Hex()
	}

	json.NewEncoder(w).Encode(&req)
}

func getLatestBoard(ID string) (GameStruct, error){
	var req GameStruct
	gameID, err := bson.ObjectIDFromHex(ID);
	if err != nil {
		return req, fmt.Errorf("Wrong GameID")
	}
	filter := bson.M{"_id": gameID}
	err = gameCollection.FindOne(context.TODO(), filter).Decode(&req)
	if err != nil {
		return req, fmt.Errorf("Internal Server Error")
	}
	return req, nil;
}

func gameChecker(ID string, playerID string) (GameStruct, error){
	var req GameStruct
	gameID, err := bson.ObjectIDFromHex(ID)
	if err != nil {
		return req, fmt.Errorf("Wrong GameID")
	}
	filter := bson.M{"_id": gameID}
	err = gameCollection.FindOne(context.TODO(), filter).Decode(&req)
	if err != nil {
		return req, fmt.Errorf("Internal Server Error")
	}
	switch{
		case req.PlayerX == "" && req.PlayerO == "":
			req.PlayerX = playerID
		case req.PlayerX != "" && req.PlayerO == "":
			req.PlayerO = playerID
		default:
			return req, fmt.Errorf("The Game is already full");
	}
	updates := bson.M{
		"$set": bson.M{
			"playerx" : req.PlayerX,
			"playero" : req.PlayerO,
		},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	gameCollection.FindOneAndUpdate(context.TODO(), filter, updates, opts).Decode(&req);
	return req, nil;
}

func play(req GameStruct, playerID string) (GameStruct, error){
	if(req.Win == false && (req.Turn == "X" && req.PlayerX != playerID || req.Turn == "O" && req.PlayerO != playerID)){
		return req, fmt.Errorf("Wait for your turn, motherfucker");
	}
	updatedGame := req;

	//Error Handling
	if req.Win == true {
		return updatedGame, fmt.Errorf("Already Won, Reset to play again")
	}
	if req.Tie == true {
		return updatedGame, fmt.Errorf("Its a tie, Reset to play again")
	}
	if req.Choice <= 0 || req.Choice > 9 {
		return updatedGame, fmt.Errorf("Wrong choice, choose between 1-9")
	}

	//Game Logic
	row := (req.Choice - 1) / 3
	column := (req.Choice - 1) % 3
	if req.Board[row][column] != "" {
		return updatedGame, fmt.Errorf("Wrong choice, spot taken")
	} else {
		updatedGame.Board[row][column] = req.Turn
		updatedGame.TurnNumber++
	}
	if checkWin(updatedGame.Board) {
		updatedGame.Win = true
	} else {
		if req.Turn == "X" {
			updatedGame.Turn = "O"
		} else {
			updatedGame.Turn = "X"
		}
	}
	if(updatedGame.TurnNumber >= 9 && updatedGame.Win == false){
		updatedGame.Tie = true;
	}

	//Encoding Data
	objID, err := bson.ObjectIDFromHex(req.GameID)
	if err != nil {
		return updatedGame, fmt.Errorf("Invalid Game ID Format")
	}
	filter := bson.M{"_id": objID}
	board := fmt.Sprintf("board.%d.%d", row, column)
	update := bson.M{
		"$set": bson.M{
			board:        req.Turn,
			"turn":       updatedGame.Turn,
			"win" :       updatedGame.Win,
			"tie" : 	  updatedGame.Tie,
			"turnNumber": updatedGame.TurnNumber,
		},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err = gameCollection.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&updatedGame)
	if err != nil {
		return updatedGame, fmt.Errorf("Database Error")
	}
	return updatedGame, nil;
}

func checkWin(board [3][3]string) bool {
	for i := range 3 {
		if (board[i][0] == board[i][1] && board[i][1] == board[i][2] && board[i][0] != "") ||
			(board[0][i] == board[1][i] && board[1][i] == board[2][i] && board[0][i] != "") {
			return true
		}
	}
	if (board[0][0] == board[1][1] && board[1][1] == board[2][2] && board[0][0] != "") ||
		(board[0][2] == board[1][1] && board[1][1] == board[2][0] && board[0][2] != "") {
		return true
	}
	return false
}

func deleteGame(ID string) (map[string]string, error){
	objID, err := bson.ObjectIDFromHex(ID)
	if err != nil {
		return nil, fmt.Errorf("Wrong GameID format")
	}
	filter := bson.M{"_id": objID}
	result, err := gameCollection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return nil, fmt.Errorf("Database Error")
	}
	if result.DeletedCount == 0 {
		return nil, fmt.Errorf("Could not find Game");
	}

	return map[string]string{"message": "Game Ended Successfully!"}, nil
}

func deletePlayer(ID string, playerID string) (GameStruct, error){
	currentBoard, err := getLatestBoard(ID);
	if err != nil{
		return currentBoard, fmt.Errorf("%v", err);
	}
	if(currentBoard.PlayerX == playerID){
		currentBoard.PlayerX = "";
		currentBoard.Win = true;
		currentBoard.Turn = "O";
	}else{
		currentBoard.PlayerO = "";
		currentBoard.Win = true;
		currentBoard.Turn = "X"
	}
	gameID, err := bson.ObjectIDFromHex(ID);
	if err != nil{
		return currentBoard, fmt.Errorf("Invalid Game ID Format");
	}
	filter := bson.M{"_id": gameID};
	updates := bson.M{
		"$set": bson.M{
			"playerx" :currentBoard.PlayerX,
			"playero" : currentBoard.PlayerO,
			"turn" : currentBoard.Turn,
			"win" : currentBoard.Win,
		},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	err = gameCollection.FindOneAndUpdate(context.TODO(), filter, updates, opts).Decode(&currentBoard)
	if err != nil{
		return currentBoard, fmt.Errorf("Database Error")
	}
	return currentBoard, nil;
}

// func reset(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "use POST Method", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	req := GameStruct{
// 		Turn: "X",
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(&req)
// }