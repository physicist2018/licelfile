package licelformat

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var licelFilenameRegex = regexp.MustCompile(`^b.*\..+`)

// LicelPack — коллекция LICEL-файлов (измерений одной сессии)
type LicelPack struct {
	StartTime time.Time            `bson:"start_time"`
	Data      map[string]LicelFile `bson:"data"`
}

func isValidFilename(filename string) bool {
	return licelFilenameRegex.MatchString(filename)
}

// NewLicelPack — загружает файлы по glob-маске
func NewLicelPack(mask string) (*LicelPack, error) {
	pack := &LicelPack{
		Data: make(map[string]LicelFile),
	}
	files, err := filepath.Glob(mask)
	if err != nil {
		return nil, fmt.Errorf("glob %q: %w", mask, err)
	}

	for i, fname := range files {
		lf, err := LoadLicelFile(fname)
		if err != nil {
			return nil, fmt.Errorf("loading %q: %w", fname, err)
		}
		pack.Data[fname] = lf
		if i == 0 {
			pack.StartTime = lf.MeasurementStartTime
		}
	}
	return pack, nil
}

// NewLicelPackFromZip — загружает файлы из zip-архива
func NewLicelPackFromZip(zipPath string) (*LicelPack, error) {
	pack := &LicelPack{
		Data: make(map[string]LicelFile),
	}
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("opening zip %q: %w", zipPath, err)
	}
	defer zr.Close()

	for _, f := range zr.File {
		if !isValidFilename(f.Name) {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("opening %q in zip: %w", f.Name, err)
		}

		fileContent, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("reading %q from zip: %w", f.Name, err)
		}

		lFile, err := LoadLicelFileFromReader(bytes.NewReader(fileContent))
		if err != nil {
			return nil, fmt.Errorf("parsing %q from zip: %w", f.Name, err)
		}

		fullPath := filepath.Join("/", f.Name)
		pack.Data[fullPath] = lFile

		if len(pack.Data) == 1 {
			pack.StartTime = lFile.MeasurementStartTime
		}
	}

	return pack, nil
}

// SelectProfiles — выбирает профили с заданной длиной волны и типом из всех файлов пака
func (lp *LicelPack) SelectProfiles(isPhoton bool, wavelength float64) LicelProfilesList {
	var result LicelProfilesList
	for _, file := range lp.Data {
		profile, ok := file.SelectProfile(isPhoton, wavelength)
		if ok {
			result = append(result, profile)
		}
	}
	return result
}

// Save — сохраняет все файлы LicelPack на диск
func (lp *LicelPack) Save() error {
	for fname, licf := range lp.Data {
		if err := licf.Save(fname); err != nil {
			return fmt.Errorf("saving %q: %w", fname, err)
		}
	}
	return nil
}

// SaveToZip — сохраняет все файлы LicelPack в zip-архив
func (lp *LicelPack) SaveToZip(zipPath string) error {
	file, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("creating zip %q: %w", zipPath, err)
	}
	defer file.Close()

	zw := zip.NewWriter(file)

	for fname, licf := range lp.Data {
		entryName := filepath.Base(fname)
		w, err := zw.Create(entryName)
		if err != nil {
			return fmt.Errorf("creating zip entry %q: %w", entryName, err)
		}
		if err := licf.WriteTo(w, entryName); err != nil {
			return fmt.Errorf("writing %q to zip: %w", entryName, err)
		}
	}

	return zw.Close()
}
