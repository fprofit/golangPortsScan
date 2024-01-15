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

type PortStatus struct {
	Port int
	Open bool
}

func scanPort(hostname string, port int, wg *sync.WaitGroup, resultChan chan PortStatus) {
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

	resultChan <- portStatus
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
	var openPortsMutex sync.Mutex

	totalPorts := 65535
	openPortsList := make([]int, 0, totalPorts)
	resultChan := make(chan PortStatus)

	for port := 1; port <= totalPorts; port++ {
		go scanPort(hostname, port, &wg, resultChan)
	}

	go func() {
		scannedPorts := make(map[int]struct{})
		for portStatus := range resultChan {
			openPortsMutex.Lock()
			if portStatus.Open {
				openPortsList = append(openPortsList, portStatus.Port)
			}
			scannedPorts[portStatus.Port] = struct{}{}
			fmt.Printf("\rProgress: %d/%d ports scanned", len(scannedPorts), totalPorts)
			openPortsMutex.Unlock()
		}
	}()

	wg.Wait()
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
