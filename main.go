package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/hpcloud/tail"
)

type Todo struct {
	ID        string `json:"id"`
	Task      string `json:"task"`
	Completed bool   `json:"completed"`
}

var todos []Todo

func GetTodoEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	for _, item := range todos {
		if item.ID == params["id"] {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	json.NewEncoder(w).Encode(&Todo{})
}

func GetTodosEndpoint(w http.ResponseWriter, req *http.Request) {
	json.NewEncoder(w).Encode(todos)
}

func CreateTodoEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	var todo Todo
	_ = json.NewDecoder(req.Body).Decode(&todo)
	todo.ID = params["id"]
	todos = append(todos, todo)
	json.NewEncoder(w).Encode(todos)
}

func DeleteTodoEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	for index, item := range todos {
		if item.ID == params["id"] {
			todos = append(todos[:index], todos[index+1:]...)
			break
		}
	}
	json.NewEncoder(w).Encode(todos)
}

type Client struct {
	socket *websocket.Conn
	send   chan []byte
}

type Broadcaster struct {
	clients    map[*Client]bool
	broadcast  chan string
	register   chan *Client
	unregister chan *Client
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func newBroadcaster() *Broadcaster {
	return &Broadcaster{
		broadcast:  make(chan string),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (b *Broadcaster) run() {
	for {
		select {
		case client := <-b.register:
			b.clients[client] = true
		case client := <-b.unregister:
			if _, ok := b.clients[client]; ok {
				delete(b.clients, client)
				close(client.send)
			}
		case message := <-b.broadcast:
			for client := range b.clients {
				client.send <- []byte(message)
			}
		}
	}
}

func readLastNLines(fileName string, n int) ([]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := make([]string, 0)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:]
		}
	}

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	return lines, nil
}

func (b *Broadcaster) initialRead(client *Client, filePath string, n int) {
	// Send last n lines from file to the client
	lines, err := readLastNLines(filePath, n)
	if err != nil {
		log.Println(err)
		return
	}
	client.send <- []byte(strings.Join(lines, "\n"))
}

func handleWebSocketConnection(b *Broadcaster, filePath string, n int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		client := &Client{socket: ws, send: make(chan []byte)}
		b.register <- client

		go b.initialRead(client, filePath, n)

		go func() {
			defer func() {
				b.unregister <- client
				ws.Close()
			}()

			for {
				_, _, err := ws.ReadMessage()
				if err != nil {
					b.unregister <- client
					ws.Close()
					break
				}
			}
		}()

		go func() {
			defer ws.Close()
			for {
				message, ok := <-client.send
				if !ok {
					ws.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				ws.WriteMessage(websocket.TextMessage, message)
			}
		}()
	}
}

func (b *Broadcaster) tailFile(filepath string) {
	t, err := tail.TailFile(
		filepath,
		tail.Config{Follow: true, Location: &tail.SeekInfo{Offset: 0, Whence: 2}},
	)
	if err != nil {
		log.Fatalf("tail file err: %v", err)
	}

	for line := range t.Lines {
		if line.Text != "" {
			b.broadcast <- line.Text
		}
	}
}

func main() {
	targetFile := "./sample_data/sample_log.log"
	lastNLines := 20

	broadcaster := newBroadcaster()
	go broadcaster.run()
	go broadcaster.tailFile(targetFile)

	staticServer := http.FileServer(http.Dir("./public_html"))
	router := mux.NewRouter()
	todos = append(todos, Todo{ID: "1", Task: "Write a medium blog post", Completed: false})
	todos = append(todos, Todo{ID: "2", Task: "Host a Go meetup", Completed: false})
	router.Handle("/", staticServer)
	router.HandleFunc("/ws", handleWebSocketConnection(broadcaster, targetFile, lastNLines))
	router.HandleFunc("/todos", GetTodosEndpoint).Methods("GET")
	router.HandleFunc("/todos", GetTodosEndpoint).Methods("GET")
	router.HandleFunc("/todos/{id}", GetTodoEndpoint).Methods("GET")
	router.HandleFunc("/todos/{id}", CreateTodoEndpoint).Methods("POST")
	router.HandleFunc("/todos/{id}", DeleteTodoEndpoint).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8000", router))
}
