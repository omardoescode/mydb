package engine

import "fmt"

type Engine struct {
}

func New() *Engine {
	return &Engine{}
}

func (eng Engine) Run() {
	fmt.Println("Hello, mydb")
}
