package main

import (
	"container/list"
	"flag"
	"golang.org/x/net/websocket"
	"io"
	"log"
	"net/http"
)

var (
	wslisten   string
	listen   string
)

func init() {
	flag.StringVar(&listen, "addr", "127.0.0.1:4000", "server address")
	flag.StringVar(&wslisten, "wsaddr", "server.net", "websocket server address")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	app := App(wslisten)
	w.Write([]byte(app))
}

type wsserver struct {
	srv *WSServer
	h   interface{}
}

type wsHandler interface {
	AddPeer(Conn *websocket.Conn) *list.Element
	LenWSPool() int
	GetPool() *list.List
	PullPeer(element *list.Element)
}

func (s wsserver) chatHandler(ws *websocket.Conn) {
	var element *list.Element

	defer func() {
		log.Print("A user just left")
		ws.Close()
		if h, ok := s.h.(wsHandler); ok {
			h.PullPeer(element)
		}
	}()

	if h, ok := s.h.(wsHandler); ok {
		element = h.AddPeer(ws)
		if element == nil {
			// There is no space for a new connection!
			websocket.Message.Send(ws, []byte("We are super busy now!"))
			return
		}
		log.Printf("We have %d users\n", h.LenWSPool())
	}

	for {
		var buf string
		err := websocket.Message.Receive(ws, &buf)
		if err != nil {
			if err == io.EOF {
				// SIGPIPE occurred.
				break
			}
			// This is a legit error so let's raise.
			panic(err)
		}
		if h, ok := s.h.(wsHandler); ok {
			pool := h.GetPool()
			for e := pool.Front(); e != nil; e = e.Next() {
				wsconn, _ := e.Value.(*websocket.Conn)
				err = websocket.Message.Send(wsconn, buf)
				if err != nil {
					// SIGPIPE occurred.
					// Get out of this loop and close
					// the connection.
					return
				}
			}
		}
	}
}

func runServer(listen string, ch chan error) {
	s := wsserver{
		srv: &WSServer{
			list.New(),
		},
	}
	s.h = s.srv
	http.HandleFunc("/", indexHandler)
	http.Handle("/chat-room", websocket.Handler(s.chatHandler))
	log.Print("HttpServer listening on ", listen)
	ch <- http.ListenAndServe(listen, nil)
}

func main() {
	flag.Parse()
	errCh := make(chan error, 1)
	go runServer(listen, errCh)
	for e := range errCh {
		if e != nil {
			panic(e)
		}
	}
}
