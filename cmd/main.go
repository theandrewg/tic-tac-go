package main

import (
	"html/template"
	"log"
	"net/http"

	tictacgo "github.com/theandrewg/tic-tac-go/internal"
)

func main() {
	game := tictacgo.NewGame()
	go game.Run()

	fs := http.FileServer(http.Dir("../css/"))
	http.Handle("/css/*", http.StripPrefix("/css/", fs))

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		t, err := template.ParseFiles("../views/index.html")
		if err != nil {
			log.Fatal(err)
		}

		data := struct {
			Players int
			Error   string
		}{
			Players: len(game.Players),
			Error:   "",
		}

		err = t.ExecuteTemplate(w, "index", data)
		if err != nil {
			log.Fatal(err)
		}
	})

	http.HandleFunc("/game", func(w http.ResponseWriter, req *http.Request) {
		game.Connect(w, req)
	})

	http.HandleFunc("/connect", func(w http.ResponseWriter, req *http.Request) {
		t, err := template.ParseFiles("../views/index.html")
		if err != nil {
			log.Fatal(err)
		}

		err = t.ExecuteTemplate(w, "connected-game", nil)
		err = t.ExecuteTemplate(w, "game-err", nil)
		if err != nil {
			log.Fatal(err)
		}
	})

	http.ListenAndServe(":42069", nil)
}
