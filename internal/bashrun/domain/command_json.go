package domain

type CommandFromUser struct {
	Command string `json:"command"`
}

type CommandFromDB struct {
	ID         int    `json:"command_id"`
	Command    string `json:"command"`
	PID        int    `json:"pid"`
	Output     string `json:"output"`
	Status     string `json:"status"`
	ExitStatus *int   `json:"exitStatus"`
}
