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
	Turn       int
}

type Box struct {
	Id     int
	Player int
}

func NewGame() *Game {
	return &Game{
		Players:    make(map[*Player]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Player),
		Unregister: make(chan *Player),
		Boxes:      initBoxes(),
		Turn:       0,
	}
}

func initBoxes() [9]Box {
	boxes := make([]Box, 0)
	for i := 0; i < 9; i++ {
		boxes = append(boxes, Box{i, 0})
	}
	return [9]Box(boxes)
}

func (g *Game) finished() bool {
	return false
}

func (g *Game) updateState() {
	if len(g.Players) == 2 && g.Turn == 0 {
		g.Turn = 1
	}
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

	players := make([]Player, 0)
	for p := range g.Players {
		players = append(players, *p)
	}

	err = t.ExecuteTemplate(&buf, "players", players)
	if err != nil {
		log.Println(err)
		return
	}

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

			t, err := template.ParseFiles("../views/index.html")
			if err != nil {
				log.Fatal(err)
			}

			var buf bytes.Buffer

			if len(g.Players) >= 2 {
				// only two players are allowed to register

				data := struct {
					Error string
				}{
					Error: "Game is full",
				}

				err = t.ExecuteTemplate(&buf, "disconnected-game", nil)
				err = t.ExecuteTemplate(&buf, "game-err", data)
				if err != nil {
					log.Fatal(err)
				}
				p.Send <- buf.Bytes()

				close(p.Send)
				break
			}

			// register the player
			g.Players[p] = true
			p.Id = len(g.Players)

			err = t.ExecuteTemplate(&buf, "disconnect", nil)
			if err != nil {
				log.Fatal(err)
			}
			p.Send <- buf.Bytes()

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
