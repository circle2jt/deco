package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/circle2jt/deco"
)

func main() {
	c := deco.New("192.168.11.1")
	err := c.Authenticate("cenpix-boVwov-4hotcj")
	if err != nil {
		log.Fatal(err.Error())
	}

	// printPerformance(c)
	var wg sync.WaitGroup

	for {
		fmt.Println("SCAN NOW")
		wg.Add(5)
		go func() {
			defer wg.Done()
			printDevices(c, "74-DA-88-5F-8F-58")
		}()
		go func() {
			defer wg.Done()
			printDevices(c, "98-DA-C4-10-5A-B0")
		}()
		go func() {
			defer wg.Done()
			printDevices(c, "74-DA-88-5F-8F-C4")
		}()
		go func() {
			defer wg.Done()
			printDevices(c, "98-DA-C4-10-52-6C")
		}()
		go func() {
			defer wg.Done()
			printDevices(c, "98-DA-C4-10-92-94")
		}()
		wg.Wait()

		time.Sleep(5 * time.Second)
	}
	// printDecos(c)
}

func printPerformance(c *deco.Client) {
	fmt.Println("[+] Permormance")
	result, err := c.Performance()
	if err != nil {
		log.Fatal(err.Error())
	}
	// Print response as json
	jsonData, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(jsonData))
}

func printDevices(c *deco.Client, decoMac string) {
	fmt.Println("[+] Clients")
	result, err := c.ClientList(decoMac)
	if err != nil {
		log.Fatal(err.Error())
	}
	if result != nil {
		for _, device := range result.Result.ClientList {
			fmt.Printf("%s\tOnline: %t\n", device.Name, device.Online)
		}
	} else {
		fmt.Println("ERROR HERE")
	}
}

func printDecos(c *deco.Client) {
	fmt.Println("[+] Devices")
	result, err := c.DeviceList()
	if err != nil {
		log.Fatal(err.Error())
	}
	for _, device := range result.Result.DeviceList {
		fmt.Printf("%s\tStatus: %s\n", device.DeviceIP, device.InetStatus)
	}
}
