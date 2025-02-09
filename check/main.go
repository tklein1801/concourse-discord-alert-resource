package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/tklein1801/concourse-discord-alert-resource/concourse"
)

func main() {
	err := json.NewEncoder(os.Stdout).Encode(concourse.CheckResponse{})
	if err != nil {
		log.Fatalln(fmt.Errorf("error: %s", err))
	}
}
