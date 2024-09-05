package server

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/pscherz/cc/data"
)

type CCClient = struct {
	Os   string
	Name string
	conn *websocket.Conn
}

var (
	clients     []CCClient
	overviewtpl *template.Template
	upgrader    = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func Init(address string) {
	const overviewhtml = `<html><body>
	List of clients:
	<ul>
	{{ range . }}
	 <li>{{.Name}} ({{.Os}})</li>
	{{ end }}
	 </ul>
	</body></html>`
	overviewtpl = template.Must(template.New("overview").Parse(overviewhtml))

	http.HandleFunc("/", handle_root)
	http.HandleFunc("/run/{id}/{cmd}", handle_run)
	http.HandleFunc("/cc/ep", handle_websocket)
	log.Printf("Server listening on %q...", address)
	log.Fatal(http.ListenAndServe(address, nil))
}

func handle_root(w http.ResponseWriter, r *http.Request) {
	overviewtpl.Execute(w, clients)
}

func handle_run(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 0 || id >= len(clients) {
		log.Printf("error parsing conn id: %v", err)
		return
	}

	client := &clients[id]
	log.Printf("sending runcmd to %q", client.Name)
	runcmd, _ := json.Marshal(data.RunCmd{Cmd: r.PathValue("cmd")})
	client.conn.WriteJSON(data.Message{
		Content: string(runcmd),
		Type:    "run",
	})
}

func handle_websocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	var msg data.Message
	err = conn.ReadJSON(&msg)
	if err != nil {
		log.Println(err)
		return
	}
	if msg.Type == "init" {
		var initmsg data.InitMessage
		err = json.Unmarshal([]byte(msg.Content), &initmsg)
		if err != nil {
			log.Println(err)
			return
		}
		clients = append(clients, CCClient{Os: initmsg.Os, Name: initmsg.Name, conn: conn})
		log.Printf("have clients: %v", len(clients))
		go handleClient(&clients[len(clients)-1])
	}
}

func handleClient(client *CCClient) {
	defer client.conn.Close()
	defer removefromclientlist(client)

	log.Print("Handling new client")

	for {
		var msg data.Message
		err := client.conn.ReadJSON(&msg)
		if err != nil {
			log.Println(client.Name, err)
			return
		}
		if msg.Type == "run" {
			// TODO: what. i actually want the output in the
			//       http handler method, how to?
			var runcmd data.RunCmd
			json.Unmarshal([]byte(msg.Content), &runcmd)
			log.Printf("cmd output (%q): %v", runcmd.Cmd, runcmd.Output)
		}
	}
}

func removefromclientlist(client *CCClient) {
	clientid := -1
	for id, c := range clients {
		if c == *client {
			clientid = id
			break
		}
	}

	if clientid < 0 {
		log.Println("did not remove client from list")
		return
	}
	log.Printf("removing client %v", clientid)
	if clientid == 0 {
		clients = clients[1:]
	} else if clientid == len(clients)-1 {
		clients = clients[0 : clientid-1]
	} else {
		clients = append(clients[0:clientid-1], clients[clientid+1:]...)
	}

}
