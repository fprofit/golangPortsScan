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

const (
	totalPorts    = 65535
	maxConcurrent = 5000
)

type PortStatus struct {
	Port int
	Open bool
}

func scanPort(hostname string, port int, sem chan struct{}, resultScanChan chan PortStatus) {
	sem <- struct{}{}
	defer func() { <-sem }()

	var portStatus PortStatus
	portStatus.Port = port

	address := fmt.Sprintf("%s:%d", hostname, port)
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		portStatus.Open = false
	} else {
		portStatus.Open = true
		conn.Close()
	}

	resultScanChan <- portStatus
}

func printStatusAndGetOpenPortsList(wg *sync.WaitGroup, resultScanChan chan PortStatus, resultChan chan []int) {
	var openPortsMutex sync.Mutex
	openPortsList := make([]int, 0, totalPorts)
	scannedPorts := 0
	for portStatus := range resultScanChan {
		openPortsMutex.Lock()
		scannedPorts++
		if portStatus.Open {
			openPortsList = append(openPortsList, portStatus.Port)
		}
		fmt.Printf("\rProgress: %d/%d ports scanned", scannedPorts, totalPorts)
		openPortsMutex.Unlock()
		wg.Done()
	}
	resultChan <- openPortsList
}

func GetOpenPorts(hostname string) []int {
	var wg sync.WaitGroup

	sem := make(chan struct{}, maxConcurrent)
	defer close(sem)

	resultChan := make(chan []int, 1)
	defer close(resultChan)

	resultScanChan := make(chan PortStatus, totalPorts)

	go printStatusAndGetOpenPortsList(&wg, resultScanChan, resultChan)

	for port := 1; port <= totalPorts; port++ {
		wg.Add(1)
		go scanPort(hostname, port, sem, resultScanChan)
	}

	wg.Wait()
	close(resultScanChan)

	openPortsList, _ := <-resultChan

	sort.Ints(openPortsList)
	return openPortsList
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

func main() {
	args := os.Args
	if !(len(args) > 1 && isValidHostname(args[1])) {
		fmt.Println("Please provide a valid hostname.")
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
