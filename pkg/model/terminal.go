package model

type Terminal struct {
	Name           string `json:"name"`
	Comment        string `json:"comment"`
	ServiceAccount struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		AccessKey AccessKey `json:"access_key"`
	} `json:"service_account"`
}

type TerminalTask struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Args       string `json:"args"`
	IsFinished bool
}
