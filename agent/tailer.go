package main

import (
	"log"

	"github.com/nxadm/tail"
)

type rawLine struct {
	service string
	text    string
}

// tailFile watches a single log file and sends each new line, tagged with
// its service name, to the lines channel.
func tailFile(service, path string, lines chan<- rawLine) {
	t, err := tail.TailFile(path, tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: true,
		Poll:      true,
	})
	if err != nil {
		log.Printf("failed to tail %s: %v", path, err)
		return
	}

	for line := range t.Lines {
		lines <- rawLine{service: service, text: line.Text}
	}
}
