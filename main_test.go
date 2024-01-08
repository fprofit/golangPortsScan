package main

import (
    "sync"
    "testing"
)

func TestScanPort(t *testing.T) {
    tests := []struct {
        hostname string
        port     int
        expected bool
    }{
        {"example.com", 80, true},
        {"localhost", 12345, false},
        {"invalid-hostname", 80, false},
    }

    for _, tt := range tests {
        wg := &sync.WaitGroup{}
        resultChan := make(chan PortStatus)

        go scanPort(tt.hostname, tt.port, wg, resultChan)
        wg.Wait()

        got := <-resultChan
        if got.Port != tt.port || got.Open != tt.expected {
            t.Errorf("scanPort(%s, %d) = %v, expected %v", tt.hostname, tt.port, got, tt.expected)
        }
        close(resultChan)
    }
}

func TestIsValidHostname(t *testing.T) {
    tests := []struct {
        hostname string
        expected bool
    }{
        {"localhost", true},
        {"127.0.0.1", true},
        {"example.com", true},
        {"invalid-hostname", false},
    }

    for _, tt := range tests {
        got := isValidHostname(tt.hostname)
        if got != tt.expected {
            t.Errorf("isValidHostname(%s) = %t, expected %t", tt.hostname, got, tt.expected)
        }
    }
}
