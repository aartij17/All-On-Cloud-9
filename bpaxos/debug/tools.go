package debug

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// For debug purposes
func WriteToFile(content string) {
	f, err := os.OpenFile("statuslog.txt",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}

	if _, err := f.WriteString(content + "\n"); err != nil {
		log.Println(err)
	}
	f.Close()
}
