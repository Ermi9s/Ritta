package ui

import (
	_ "embed"
	"fmt"
)

//go:embed ritta.ansi
var rittaANSI []byte

func Banner() string {
	return string(rittaANSI)
}

func PrintBanner() {
	fmt.Print(Banner())
	fmt.Print()
}