package main

import "github.com/schidstorm/sshd-manager/http"

func main() {
	err := http.Run("localhost:8080")
	if err != nil {
		panic(err)
	}
}

