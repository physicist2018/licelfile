package main

import (
	"fmt"

	"github.com/physicist2018/licelfile/licelformat"
)

func main() {

	a := licelformat.LoadLicelFile("a16A0421.070296")
	for i := range a.Profiles {
		println(a.Profiles[i].Wavelength, a.Profiles[i].Photon, a.Profiles[i].NShots, a.Profiles[i].DiscrLevel)
		fmt.Printf("%.2f %.2f %.2f\n", a.Profiles[i].Data[0], a.Profiles[i].Data[1], a.Profiles[i].Data[2])

	}
	//_ = licelformat.NewLicelPackFromZip("111.zip")

	//w := licelformat.SelectCertainWavelength2(&a, false, 355)
	//_ = json.NewEncoder(os.Stdout).Encode(a)
	//a.Save()
	//for _, v := range v[0].Data {
	//	fmt.Println(v)
	//}

}
