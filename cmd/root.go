package cmd

import "log"

// Execute starts the HTTP application.
func Execute() {
	if err := runHTTP(); err != nil {
		log.Fatal(err)
	}
}
