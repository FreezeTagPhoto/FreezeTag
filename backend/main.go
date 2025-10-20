//go:build !test

package main

import (
	"fmt"
)

func main() {
	fmt.Println(Hello("World"))
}
