package types

type Task struct {
	HandlerName string `json:"handler_name" yaml:"handler_name"`
	Data        []byte `json:"data" yaml:"data"`
}

type Schedule []Task
