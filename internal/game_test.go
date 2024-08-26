package tictacgo_test

import (
	"testing"

	"github.com/gorilla/websocket"
	tictacgo "github.com/theandrewg/tic-tac-go/internal"
)

func Test_RegisterPlayer(t *testing.T) {
	g := tictacgo.NewGame()
	go g.Run()

    conn := &websocket.Conn{}
	p := &tictacgo.Player{
		Game: g,
		Conn: conn,
		Send: make(chan []byte, 256),
	}
	g.Register <- p
}

func Test_RegisterThreePlayers(t *testing.T) {
	g := tictacgo.NewGame()
	go g.Run()

	c1 := &websocket.Conn{}
	p1 := &tictacgo.Player{
		Game: g,
		Conn: c1,
		Send: make(chan []byte, 256),
	}
	g.Register <- p1

	c2 := &websocket.Conn{}
	p2 := &tictacgo.Player{
		Game: g,
		Conn: c2,
		Send: make(chan []byte, 256),
	}
	g.Register <- p2

    c3 := &websocket.Conn{}
	p3 := &tictacgo.Player{
		Game: g,
		Conn: c3,
		Send: make(chan []byte, 256),
	}
	g.Register <- p3

    if len(g.Players) != 2 {
        t.Errorf("There are %d players registered. There should be 2", len(g.Players))
    }
}
