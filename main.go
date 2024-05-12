package main

import (
	// "golang.org/x/crypto/ssh"
	"fmt"
	"sync"

	"github.com/tdavari/sshapp/repository/ssh"
)

func main() {
	var devices []ssh.Device
	devices, _ = ssh.Load("test")

	fmt.Printf("%+v\n", devices)

	// ssh.ConnectViaJump(devices[1:], devices[0])
	jumpbox, err := devices[0].Connect()
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	for _, d := range devices[2:] {

		wg.Add(1)
		go func() {
			defer wg.Done()
			sw, err := d.ConnectViaJump(jumpbox)
			if err != nil {
				panic(err)
			}
			output, err := ssh.ExecuteCommand(sw, "show clock")
			if err != nil {
				panic(err)
			}
			fmt.Println(output)
		}()
	}
	wg.Wait()
}
