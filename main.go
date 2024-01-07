package main

import (
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

func scanPort(hostname string, port int, wg *sync.WaitGroup, progress, openPorts chan int) {
	defer wg.Done()
	progress <- port
	address := fmt.Sprintf("%s:%d", hostname, port)
	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		return
	}
	defer conn.Close()
	openPorts <- port
}

func printProgres(totalPorts int, progress chan int) {
	var progressMutex sync.Mutex
	scannedPorts := make(map[int]struct{})
	for port := range progress {
		progressMutex.Lock()
		scannedPorts[port] = struct{}{}
		fmt.Printf("\rProgress: %d/%d ports scanned", len(scannedPorts), totalPorts)
		progressMutex.Unlock()
	}
}

func isValidHostname(hostname string) bool {
	if ip := net.ParseIP(hostname); ip != nil {
		return true
	}
	if _, err := net.LookupHost(hostname); err == nil {
		return true
	}
	return false
}

func GetOpenPorts(hostname string) []int {
	var wg sync.WaitGroup
	var openPortsListMutex sync.Mutex

	totalPorts := 65535
	openPortsList := make([]int, 0, totalPorts)

	wg.Add(totalPorts)
	progress := make(chan int)
	openPorts := make(chan int)

	for port := 1; port <= totalPorts; port++ {
		go scanPort(hostname, port, &wg, progress, openPorts)
	}
	go printProgres(totalPorts, progress)
	go func() {
		for port := range openPorts {
			openPortsListMutex.Lock()
			openPortsList = append(openPortsList, port)
			openPortsListMutex.Unlock()
		}
	}()

	wg.Wait()
	close(progress)
	close(openPorts)

	sort.Ints(openPortsList)
	return openPortsList
}

func main() {
	args := os.Args
	if !(len(args) > 1 && isValidHostname(args[1])) {
		fmt.Println("The hostname argument is missing or invalid.")
		return
	}

	openPortsList := GetOpenPorts(args[1])

	fmt.Println("\nScan completed.")
	if len(openPortsList) == 0 {
		fmt.Println("No open ports")
	} else {
		openPortsString := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(openPortsList)), ", "), "[]")
		fmt.Println("Open ports:", openPortsString)
	}
}
