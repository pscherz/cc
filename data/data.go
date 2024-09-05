package data

type Message = struct {
	Content string `json:"content"`
	Type    string `json:"type"`
}

type InitMessage = struct {
	Os   string `json:"os"`
	Name string `json:"name"`
}

type RunCmd = struct {
	Cmd    string
	Output string
}
