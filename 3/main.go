package main

import (
	"fmt"
)

type User struct {
	Name string
}

func (u *User) Hello() {
	fmt.Println("Hello " + u.Name)
}

type Hi interface {
	Hello()
}

func main() {

}
