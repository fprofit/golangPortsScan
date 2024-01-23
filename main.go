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
	totalPorts = 65535
)

type PortStatus struct {
	Port int
	Open bool
}

func scanPort(hostname string, port int, wg *sync.WaitGroup, resultScanChan chan PortStatus) {
	wg.Add(1)
	defer wg.Done()

	var portStatus PortStatus
	portStatus.Port = port

	address := fmt.Sprintf("%s:%d", hostname, port)
	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err == nil {
		portStatus.Open = true
		if conn != nil {
			conn.Close()
		}
	}
	resultScanChan <- portStatus
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

func printStatusAndGetOpenPortsList(resultScanChan chan PortStatus, resultChan chan []int) {
	var openPortsMutex sync.Mutex
	openPortsList := make([]int, 0, totalPorts)
	scannedPorts := make(map[int]struct{}, totalPorts)
	for portStatus := range resultScanChan {
		openPortsMutex.Lock()
		if portStatus.Open {
			openPortsList = append(openPortsList, portStatus.Port)
		}
		scannedPorts[portStatus.Port] = struct{}{}
		fmt.Printf("\rProgress: %d/%d ports scanned", len(scannedPorts), totalPorts)
		openPortsMutex.Unlock()
	}
	resultChan <- openPortsList
}

func GetOpenPorts(hostname string) []int {
	var wgScan sync.WaitGroup
	resultScanChan := make(chan PortStatus, 1000)
	resultChan := make(chan []int, 1)

	for port := 1; port <= totalPorts; port++ {
		go scanPort(hostname, port, &wgScan, resultScanChan)
	}

	go printStatusAndGetOpenPortsList(resultScanChan, resultChan)

	wgScan.Wait()
	close(resultScanChan)

	openPortsList, _ := <-resultChan
	close(resultChan)

	sort.Ints(openPortsList)
	return openPortsList
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
