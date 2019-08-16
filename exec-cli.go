package main

import (
	"fmt"
	"os/exec"
	"time"
)

func main() {
	for i := 0; i < 1000; i++ {
		err := exec.Command("./client-one", "-port "+string(33333+i), "-file "+"../checkFiles/src/"+string(i+1)+".bin").Run()
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(10)
	}
}
