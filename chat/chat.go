package main

import (
	"flag"
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
)

const maxWSConn = 32

var (
	listen     string
	wslisten   string
	wsRefc     int
	// TODO:
	// This is bad because we cannot deal with closed connections.
	// The order matters too much so we should use some stuff with
	// which we can deal with the order.
	wsConnPool []*websocket.Conn
)

func init() {
	flag.StringVar(&listen, "addr", "192.168.1.1:4000", "server address")
	flag.StringVar(&wslisten, "wsaddr", "192.168.1.1:4001", "websocket server address")
	wsRefc = 0
	wsConnPool = make([]*websocket.Conn, maxWSConn)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	out := fmt.Sprintf(`
<html>
    <head>
	<meta charset="utf-8">
	<style type="text/css">
	 #container {
	     position: relative;
	     font-family: helvetica;
	 }
	 #bottom {
	     position: fixed;
	     bottom: 0px;
	 }
	 #messages {
	     padding: 0em 0em;
	 }
	 ul, li {
	     list-style: none;
	 }
	 input {
	     display: block;
	     margin-bottom: 10px;
	     padding: 5px;
	     font-size: 16px;
	     width: 500px;
	     max-width: 800px;
	     min-width: 400px;
	 }
	</style>
    </head>
    <body>
	<div id="container">
	    <div id="message-grid">
		<ul id="messages"></ul>
	    </div>
	    <div id="bottom">
		<input id="input-field" />
	    </div>
	</div>
	<script>
	 function connect (wsserver, onOpenFunc, onMessageFunc) {
	     return new Promise(function(resolve, reject) {
		 const socket = new WebSocket(wsserver);
		 socket.addEventListener('open', function() {
		     onOpenFunc(socket);
		 });
		 socket.addEventListener('message', function(evt) {
		     onMessageFunc(evt);
		 });
		 socket.onopen = function() {
		     resolve(socket);
		 };
		 socket.onerror = function(err) {
		     reject(err);
		 };
	     });
	 };
	 const onOpenFunc = function(sock) {
	     const input = document.getElementById('input-field');
	     input.addEventListener('change', function(evt) {
		 sock.send(evt.target.value);
		 evt.target.value = '';
	     });
	 };
	 const onMessageFunc = function(evt) {
	     const grid = document.getElementById('messages')
	     const data = evt.data;
	     const timestamp = new Date();
	     grid.innerHTML += '<li>'+timestamp.toLocaleString()+': '+data+'</li>';
	 };
	 connect('ws://%s/chat-room', onOpenFunc, onMessageFunc).then(function(conn) {
	     console.log('connected:', conn);
	 }).catch(function(err) {
	     throw('error:', err);
	 });
	</script>
    </body>
</html>
`, wslisten)
	w.Write([]byte(out))
}
func chatHandler(ws *websocket.Conn) {
	if wsRefc >= maxWSConn {
		ws.Write([]byte("Room is a bit crowded. Hang tight 5~ minutes til this pool is drained out."))
		log.Print("crowded")
		return
	}

	defer func() {
		log.Print("A user just left")
		ws.Close()
		wsRefc--
	}()

	// Add a new client.
	wsConnPool[wsRefc] = ws
	wsRefc++
	log.Printf("We have %d users\n", wsRefc)
	for {
		var buf string
		err := websocket.Message.Receive(ws, &buf)
		if err != nil {
			break
		}
		for i := 0; i < wsRefc; i++ {
			err = websocket.Message.Send(wsConnPool[i], buf)
			if err != nil {
				break
			}
		}
	}
}

func runWSServer(listen string, ch chan error) {
	http.Handle("/chat-room", websocket.Handler(chatHandler))
	log.Print("WebSocket server listening on ", listen)
	err := http.ListenAndServe(listen, nil)
	ch <- err
}

func runServer(listen string, ch chan error) {
	http.HandleFunc("/", indexHandler)
	log.Print("HttpServer listening on ", listen)
	err := http.ListenAndServe(listen, nil)
	ch <- err
}

func main() {
	flag.Parse()
	errCh := make(chan error, 2)
	go runServer(listen, errCh)
	go runWSServer(wslisten, errCh)
	for e := range errCh {
		if e != nil {
			panic(e)
		}
	}
}
