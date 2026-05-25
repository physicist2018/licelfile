package main

import (
	"fmt"
	"os"

	"github.com/physicist2018/licelfile/licelformat"
)

func main() {
	a, err := licelformat.LoadLicelFile("a16A0421.070296")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading file: %v\n", err)
		os.Exit(1)
	}
	for i := range a.Profiles {
		println(a.Profiles[i].Wavelength, a.Profiles[i].Photon, a.Profiles[i].NShots, a.Profiles[i].DiscrLevel)
		fmt.Printf("%.2f %.2f %.2f\n", a.Profiles[i].Data[0], a.Profiles[i].Data[1], a.Profiles[i].Data[2])
	}
}
