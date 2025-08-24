package domain

import (
	"os"

	"log"
)

var DomainLogger = log.New(os.Stderr, "DOMAIN", log.Ldate)
