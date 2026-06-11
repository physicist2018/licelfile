package main

import (
	"fmt"
	"os"

	"github.com/physicist2018/licelfile/v2/licelformat"
)

func main() {
	a, err := licelformat.NewLicelPackFromZip("archive.zip")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading file: %v\n", err)
		os.Exit(1)
	}
	err = a.SaveToNetCDF3("1.nc")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving file: %v\n", err)
		os.Exit(1)
	}
	b, err := licelformat.LoadLicelPackFromNetCDF3("1.nc")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(b.StartTime)
}
