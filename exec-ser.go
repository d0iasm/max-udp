package main

import (
	"fmt"
	"os/exec"
	"time"
)

func main() {
	i := 0
	for {
		if i > 1000 {
			i = 0
		}
		err := exec.Command("./server-one", "-port "+string(33333+i), "-file"+"../checkFiles/dst/"+string(i+1)+".bin").Run()
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(10)
		i++
	}
}
