package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	socketio "github.com/googollee/go-socket.io"
)

func hello() {
	fmt.Println("Hello Plus7")
}

func main() {
	hello()

	server := socketio.NewServer(nil)

	var socket net.Conn = nil

	server.OnConnect("/", func(client socketio.Conn) error {
		client.SetContext("")
		log.Println("connected:", client.ID())
		return nil
	})

	server.OnEvent("/", "socket_open", func(client socketio.Conn, uid int, protocol string, ip string, port int) {
		destination := ip + ":" + strconv.Itoa(port)

		log.Println("open socket:", uid, "@", destination)

		newConnection, err := net.Dial("tcp", destination)
		socket = newConnection

		if err != nil {
			log.Println("dial error:", err)

			client.Emit("socket_not_opened", uid)
			return
		}

		go (func() {
			for {
				buffer := make([]byte, 1024)

				readLen, err := newConnection.Read(buffer)

				if err != nil {
					return
				}

				log.Println("[READ]", uid, ":", readLen, "bytes")

				client.Emit("socket_read", uid, buffer[:readLen])
			}
		})()

		client.Emit("socket_opened", uid)
	})

	server.OnEvent("/", "socket_write", func(client socketio.Conn, uid int, data []byte) {
		log.Println("[WRITE]", uid, ":", len(data), "bytes")

		socket.Write(data)
	})

	server.OnEvent("/", "socket_close", func(client socketio.Conn) string {
		last := client.Context().(string)
		client.Emit("bye", last)
		client.Close()
		return last
	})

	server.OnError("/", func(client socketio.Conn, e error) {
		log.Println("meet error:", e)
	})

	server.OnDisconnect("/", func(client socketio.Conn, reason string) {
		log.Println("closed", reason)
	})

	go server.Serve()
	defer server.Close()

	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./asset")))

	log.Println("Serving at localhost:8000...")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
