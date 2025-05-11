package licelformat

import (
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
)

type LicelPack struct {
	ID        string               `bson:"_id,omitempty"`
	StartTime time.Time            `bson:"start_time"`
	Data      map[string]LicelFile `bson:"data"`
}

// NewLicelPack — loads files according to mask
func NewLicelPack(mask string) *LicelPack {
	pack := &LicelPack{
		Data: make(map[string]LicelFile),
	}
	files, err := filepath.Glob(mask)
	if err != nil {
		log.Fatal().Err(err).Str("mask", mask).Msg("Error getting files by mask")
	}
	for i, fname := range files {
		pack.Data[fname] = LoadLicelFile(fname)
		if i == 0 {
			pack.StartTime = pack.Data[fname].MeasurementStartTime
		}

		//println(pack[fname].)

	}
	return pack
}

// SelectCertainWavelength2 — selects certain profile by its wavelength and type from a LicelPack
func (lp *LicelPack) SelectCertainWavelength(isPhoton bool, wavelength float64) LicelProfilesList {
	var result LicelProfilesList
	for _, file := range lp.Data {
		profile := file.SelectCertainWavelength(isPhoton, wavelength)
		if profile.Wavelength != 0 {
			result = append(result, profile)
		}
	}
	return result
}

// Save - saves LicelPack to files
func (lp *LicelPack) Save() error {
	for fname, licf := range lp.Data {
		if err := licf.Save(fname); err != nil {
			return err
		}
	}
	return nil
}
