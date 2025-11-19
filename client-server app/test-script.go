package main

import (
	"fmt"
)

func main() {
	const s string = "Ana000!"
	hasUpper, hasLower, hasDigit, hasSymbol := false, false, false, false

	for _, ch := range s {
		switch {
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		default:
			hasSymbol = true
		}
	}

	if hasUpper && hasLower && hasDigit && hasSymbol {
		fmt.Println("accepted")
	} else {
		fmt.Println("not accepted")
	}

}
