package licelformat

import (
	"archive/zip"
	"bytes"
	"io"
	"path/filepath"
	"regexp"
	"time"

	"github.com/rs/zerolog/log"
)

type LicelPack struct {
	StartTime time.Time            `bson:"start_time"`
	Data      map[string]LicelFile `bson:"data"`
}

func isValidFilename(filename string) bool {
	match, _ := regexp.MatchString("^b.*\\..+", filename)
	return match
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

// NewLicelPackFromZip — loads files from zip archive
func NewLicelPackFromZip(zipPath string) *LicelPack {
	pack := &LicelPack{
		Data: make(map[string]LicelFile),
	}
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		log.Error().Err(err)
	}
	defer zr.Close()

	for _, f := range zr.File {
		if isValidFilename(f.Name) {
			log.Info().Str("Loading file ", f.Name).Send()

			rc, err := f.Open()
			if err != nil {
				log.Info().Str("Error opening file ", f.Name).Err(err).Send()
				continue
			}
			defer rc.Close()

			// Читаем файл прямо из потока, не распаковывая весь архив
			fileContent, err := io.ReadAll(rc)
			if err != nil {
				log.Info().Str("Error reading file content of ", f.Name).Err(err).Send()
				continue
			}

			// Загружаем файл с помощью LoadLicelFile
			lFile := LoadLicelFileFromReader(bytes.NewReader(fileContent), int64(len(fileContent)))

			// добавляем файл в карту данных
			fullPath := filepath.Join("/", f.Name)
			pack.Data[fullPath] = lFile

			// устанавливаем StartTime из первого файла
			if len(pack.Data) == 1 {
				pack.StartTime = lFile.MeasurementStartTime
			}
		}
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
