package tictacgo

import (
	"bytes"
	"html/template"
	"log"
)

type Game struct {
	Players    map[*Player]bool
	Broadcast  chan []byte
	Register   chan *Player
	Unregister chan *Player
}

func NewGame() *Game {
	return &Game{
		Players:    make(map[*Player]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Player),
		Unregister: make(chan *Player),
	}
}

func (g *Game) updateState() {
	t, err := template.ParseFiles("./views/index.html")
	if err != nil {
		log.Fatal(err)
	}

	data := struct {
		Count int
	}{
		Count: len(g.Players),
	}

	var buf bytes.Buffer

	err = t.ExecuteTemplate(&buf, "counter", data)
	if err != nil {
		log.Println(err)
		return
	}

	b := buf.Bytes()
	for p := range g.Players {
		p.Send <- b
	}
}

func (g *Game) Run() {
	for {
		select {
		// register a new player
		case p := <-g.Register:
			if len(g.Players) == 2 {
				// only two players are allowed to register
				close(p.Send)
			}
			g.Players[p] = true
			g.updateState()

		// unregister a player
		case p := <-g.Unregister:
			if _, ok := g.Players[p]; ok {
				close(p.Send)
				delete(g.Players, p)
			}
			g.updateState()

		// send a message to all players
		case message := <-g.Broadcast:
			for p := range g.Players {
				p.Send <- message
			}
		}
	}
}
