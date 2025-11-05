package hello

import "fmt"

func Hello(name string) string {
	if len(name) == 0 {
		return ""
	}
	message := fmt.Sprintf("Hi, %v. Welcome!", name)
	return message
}
