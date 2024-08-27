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
	Boxes      [9]Box
}

func NewGame() *Game {
	return &Game{
		Players:    make(map[*Player]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Player),
		Unregister: make(chan *Player),
		Boxes:      initBoxes(),
	}
}

type Box struct {
	Id     int
	Player int
}

func NewBox(i int) Box {
	return Box{Id: i, Player: 0}
}

func initBoxes() [9]Box {
	return [9]Box(append(make([]Box, 0), NewBox(0), NewBox(1), NewBox(2), NewBox(3), NewBox(4),
		NewBox(5), NewBox(6), NewBox(7), NewBox(8)))
}

func (g *Game) updateState() {
	t, err := template.ParseFiles("../views/index.html")
	if err != nil {
		log.Fatal(err)
	}

	data := struct {
		Count int
		Boxes [9]Box
	}{
		Count: len(g.Players),
		Boxes: g.Boxes,
	}

	var buf bytes.Buffer

	err = t.ExecuteTemplate(&buf, "boxes", data)
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
			if p == nil {
				log.Panic("unable to register nil player")
			}
			if len(g.Players) >= 2 {
				// only two players are allowed to register
				t, err := template.ParseFiles("../views/index.html")
				if err != nil {
					log.Fatal(err)
				}

				data := struct {
					Error string
				}{
					Error: "Game is full",
				}

				var buf bytes.Buffer
				err = t.ExecuteTemplate(&buf, "disconnected-game", nil)
				err = t.ExecuteTemplate(&buf, "game-err", data)
				if err != nil {
					log.Fatal(err)
				}
				p.Send <- buf.Bytes()

				close(p.Send)
			} else {
				g.Players[p] = true
                p.Id = len(g.Players)
				g.updateState()
			}

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
