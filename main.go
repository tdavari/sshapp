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
	err := devices[0].Connect()
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	for _, d := range devices[1:] {

		wg.Add(1)
		go func() {
			defer wg.Done()
			if d.IsConnected() = false{
				err := d.ConnectViaJump(devices[0].Client)
			}
			

			if err != nil {
				panic(err)
			}
			output, err := d.ExecuteCommand("show clock")
			if err != nil {
				panic(err)
			}
			fmt.Println(output)
		}()
	}
	wg.Wait()
	println(devices[1].IsConnected())
	println(devices[2].IsConnected())
	println(devices[0].IsConnected())

	for _, d := range devices[1:] {

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := d.ConnectViaJump(devices[0].Client)

			if err != nil {
				panic(err)
			}
			output, err := d.ExecuteCommand("show clock")
			if err != nil {
				panic(err)
			}
			fmt.Println(output)
		}()
	}
	wg.Wait()
	println(devices[1].IsConnected())
	println(devices[2].IsConnected())
	println(devices[0].IsConnected())
}
