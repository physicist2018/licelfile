package main

import (
	"fmt"

	"github.com/physicist2018/licelfile/licelformat"
)

func main() {

	a := licelformat.NewLicelPack("b*.*")

	v := licelformat.SelectCertainWavelength2(&a, true, 408)
	fmt.Println(v)
	for key := range a {
		fmt.Println(key)
	}
}
