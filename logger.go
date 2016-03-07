package osc

import (
	"log"
	"os"
)

var logger = log.New(os.Stdout, "[osc] ", log.Lshortfile)
