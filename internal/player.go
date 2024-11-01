package tictacgo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// message read timer
	readWait = 60 * time.Second

	// message write timer
	writeWait = 10 * time.Second

	// client ping frequency
	// must be less than readWait
	pingFreq = readWait * 9 / 10

	// max client sent message size
	msgLimit = 1024
)

const (
	Unknown int = iota
	Close
	Select
	Reset
)

var (
	newLine = []byte{'\n'}
	space   = []byte{' '}
)

type Player struct {
	Game *Game
	Conn *websocket.Conn
	Send chan []byte
	Id   int
}

type ClientMessage struct {
	Type int
}

type SelectMessage struct {
	Type   int
	Box    int
	Player int
}

func (p *Player) ReadMessages() {
	defer func() {
		p.Game.Unregister <- p
	}()

	p.Conn.SetReadLimit(msgLimit)
	p.Conn.SetReadDeadline(time.Now().Add(readWait))
	p.Conn.SetPongHandler(func(string) error {
		p.Conn.SetReadDeadline(time.Now().Add(readWait))
		return nil
	})

readLoop:
	for {
		_, msgBytes, err := p.Conn.ReadMessage()
		if err != nil {
			log.Print(err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		var msg ClientMessage

		err = json.Unmarshal(msgBytes, &msg)
		if err != nil {
			log.Fatal(err)
		}

		switch msg.Type {
		default:
		case Unknown:
			p.Send <- []byte("unknown type")
		case Select:
			if p.Id != p.Game.Turn || p.Game.Winner != 0 {
				break
			}

			var s SelectMessage
			err = json.Unmarshal(msgBytes, &s)
			if err != nil {
				log.Fatal(err)
			}

			winner := p.Game.selectBox(s.Box, p.Id)

			t, err := template.ParseFiles("../views/index.html")
			if err != nil {
				log.Fatal(err)
			}

			data := struct {
				Id     int
				Player int
			}{
				Id:     s.Box,
				Player: p.Id,
			}

			var buf bytes.Buffer

			err = t.ExecuteTemplate(&buf, "box", data)
			if err != nil {
				log.Println(err)
				return
			}

			if winner != 0 {
				err = t.ExecuteTemplate(&buf, "winner", fmt.Sprintf("Player %d wins!", winner))
				if err != nil {
					log.Println(err)
					return
				}
			}

			b := buf.Bytes()
			p.Game.Broadcast <- b
		case Close:
			p.Game.Boxes = initBoxes()
			t, err := template.ParseFiles("../views/index.html")
			if err != nil {
				log.Fatal(err)
			}

			data := struct {
				Players int
				Error   string
			}{
				Players: 0,
				Error:   "",
			}

			var buf bytes.Buffer

			err = t.ExecuteTemplate(&buf, "disconnected-game", data)
			if err != nil {
				log.Println(err)
				return
			}

			b := buf.Bytes()
			for player := range p.Game.Players {
				delete(player.Game.Players, player)
				player.Send <- b
				player.Game.Unregister <- player
			}
			break readLoop
		case Reset:
			p.Game.reset()

			t, err := template.ParseFiles("../views/index.html")
			if err != nil {
				log.Fatal(err)
			}

			data := struct {
				Count int
				Boxes [9]Box
			}{
				Count: len(p.Game.Players),
				Boxes: p.Game.Boxes,
			}

			var buf bytes.Buffer
			err = t.ExecuteTemplate(&buf, "boxes", data)
			err = t.ExecuteTemplate(&buf, "winner", "")
			if err != nil {
				log.Println(err)
				return
			}

			b := buf.Bytes()
			// for p := range p.Game.Players {
			// 	p.Send <- b
			// }
			p.Game.Broadcast <- b
		}
	}
}

func (p *Player) WriteMessages() {
	ticker := time.NewTicker(pingFreq)
	defer func() {
		ticker.Stop()
		p.Game.Unregister <- p
	}()

	for {
		select {
		case msg, ok := <-p.Send:
			p.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				p.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := p.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			w.Write(msg)

			n := len(p.Send)
			for i := 0; i < n; i++ {
				w.Write(newLine)
				w.Write(<-p.Send)
			}

			err = w.Close()
			if err != nil {
				return
			}
		case <-ticker.C:
			p.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			err := p.Conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				return
			}
		}
	}
}
