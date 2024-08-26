package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/theandrewg/tic-tac-go/internal"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func connectGame(game *tictacgo.Game, res http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Println(err)
		return
	}

	player := &tictacgo.Player{
		Game: game,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	game.Register <- player

    go player.ReadMessages()
    go player.WriteMessages()
}

func main() {
	game := tictacgo.NewGame()
	go game.Run()

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		t, err := template.ParseFiles("../views/index.html")
		if err != nil {
			log.Fatal(err)
		}

		err = t.ExecuteTemplate(w, "index", nil)
		if err != nil {
			log.Fatal(err)
		}
	})

	http.HandleFunc("/game", func(w http.ResponseWriter, req *http.Request) {
		connectGame(game, w, req)
	})

	http.HandleFunc("/connect", func(w http.ResponseWriter, req *http.Request) {
		t, err := template.ParseFiles("./views/index.html")
		if err != nil {
			log.Fatal(err)
		}

		err = t.ExecuteTemplate(w, "connected-game", nil)
		if err != nil {
			log.Fatal(err)
		}
	})

	http.ListenAndServe(":42069", nil)
}
