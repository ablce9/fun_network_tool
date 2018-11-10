package main

import "fmt"

// App is like jsx.
func App(wsserver string) string {
	return fmt.Sprintf(`<html>
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
`, wsserver)
}
