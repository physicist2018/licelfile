package licelformat

import (
	"path/filepath"

	"github.com/rs/zerolog/log"
)

type LicelPack map[string]LicelFile

// NewLicelPack — loads files according to mask
func NewLicelPack(mask string) LicelPack {
	pack := make(LicelPack)
	files, err := filepath.Glob(mask)
	if err != nil {
		log.Fatal().Err(err).Str("mask", mask).Msg("Error getting files by mask")
	}
	for _, fname := range files {
		pack[fname] = LoadLicelFile(fname)
		println(pack[fname].NDatasets)
		//println(pack[fname].)

	}
	return pack
}

// SelectCertainWavelength2 — selects certain profile by its wavelength and type from a LicelPack
func (lp *LicelPack) SelectCertainWavelength(isPhoton bool, wavelength float64) LicelProfilesList {
	var result LicelProfilesList
	for _, file := range *lp {
		profile := file.SelectCertainWavelength(isPhoton, wavelength)
		if profile.Wavelength != 0 {
			result = append(result, profile)
		}
	}
	return result
}

func (lp *LicelPack) Save() error {
	for fname, licf := range *lp {
		if err := licf.Save(fname); err != nil {
			return err
		}
	}
	return nil
}
