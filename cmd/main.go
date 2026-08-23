package main

import (
	"log"
	"ritta/internal/cli"
)


func main() {
	if err:= cli.Execute(); err != nil {
		log.Fatalf("Error %v", err);
	} 

	
}