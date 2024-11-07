package helpers

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

func GetPassword() string {
	fmt.Print("Password: ")
	p, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	fmt.Println()
	return string(p)
}
