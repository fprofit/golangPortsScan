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
        {"invalid-hostname", 1234, false},
        {"localhost", 12345, false},
    }

    for _, tt := range tests {
        func(hostname string, port int, expected bool) {
            var wg sync.WaitGroup
            progress := make(chan int)
            openPorts := make(chan int)

            wg.Add(1)
            go scanPort(hostname, port, &wg, progress, openPorts)
            go func() {
                wg.Add(1)
                defer wg.Done()
                p := <-progress
                if p != port {
                    t.Errorf("Expected progress channel to receive %d, but got %d", port, port)
                }
                if expected {
                    open := <-openPorts
                    if open != port {
                        t.Errorf("Expected openPorts channel to receive %d, but got %d", port, open)
                    }
                }
            }()
            wg.Wait()
            close(openPorts)
            close(progress)
        }(tt.hostname, tt.port, tt.expected)
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
        {"", false},
    }

    for _, tt := range tests {
        got := isValidHostname(tt.hostname)
        if got != tt.expected {
            t.Errorf("isValidHostname(%s) = %t, expected %t", tt.hostname, got, tt.expected)
        }
    }
}
