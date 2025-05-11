package main

import (
	"github.com/physicist2018/licelfile/licelformat"
)

func main() {

	a := licelformat.NewLicelPack("b*.*")

	//w := licelformat.SelectCertainWavelength2(&a, false, 355)
	//_ = json.NewEncoder(os.Stdout).Encode(a)
	a.Save()
	//for _, v := range v[0].Data {
	//	fmt.Println(v)
	//}

}
