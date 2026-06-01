package licelformat

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var licelFilenameRegex = regexp.MustCompile(`^[a-z].*\..+`)

// LicelPack — коллекция LICEL-файлов (измерений одной сессии)
type LicelPack struct {
	StartTime           time.Time            `bson:"start_time"`
	StopTime            time.Time            `bson:"stop_time"`
	Data                map[string]LicelFile `bson:"data"`
	ZipCompressionLevel int                  `bson:"-"` // 0 = default deflate, 1–9 = уровень сжатия
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

	for _, fname := range files {
		lf, err := LoadLicelFile(fname)
		if err != nil {
			return nil, fmt.Errorf("loading %q: %w", fname, err)
		}
		pack.Data[fname] = lf
	}

	var minStart, maxStop time.Time
	for _, lf := range pack.Data {
		if minStart.IsZero() || lf.MeasurementStartTime.Before(minStart) {
			minStart = lf.MeasurementStartTime
		}
		if lf.MeasurementStopTime.After(maxStop) {
			maxStop = lf.MeasurementStopTime
		}
	}

	pack.StartTime = minStart
	pack.StopTime = maxStop
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

	}

	var minStart, maxStop time.Time
	for _, lf := range pack.Data {
		if minStart.IsZero() || lf.MeasurementStartTime.Before(minStart) {
			minStart = lf.MeasurementStartTime
		}
		if lf.MeasurementStopTime.After(maxStop) {
			maxStop = lf.MeasurementStopTime
		}
	}

	pack.StartTime = minStart
	pack.StopTime = maxStop

	return pack, nil
}

// SelectProfiles — выбирает профили с заданными длиной волны, типом и поляризацией из всех файлов пака.
// Передайте "" в polarization чтобы подходила любая.
func (lp *LicelPack) SelectProfiles(isPhoton bool, wavelength float64, polarization string) LicelProfilesList {
	var result LicelProfilesList
	for _, file := range lp.Data {
		profile, ok := file.SelectProfile(isPhoton, wavelength, polarization)
		if ok {
			result = append(result, profile)
		}
	}
	return result
}

// SetMaxDist обрезает все профили во всех файлах пака до дальности alt (метры).
func (lp *LicelPack) SetMaxDist(alt float64) error {
	for fname, licf := range lp.Data {
		if err := licf.SetMaxDist(alt); err != nil {
			return fmt.Errorf("%s: %w", fname, err)
		}
		lp.Data[fname] = licf
	}
	return nil
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

// SaveToZip — сохраняет все файлы LicelPack в zip-архив.
// Уровень сжатия задаётся полем ZipCompressionLevel: 0 — deflate по умолчанию, 1–9 — степень deflate.
func (lp *LicelPack) SaveToZip(zipPath string) error {
	file, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("creating zip %q: %w", zipPath, err)
	}
	defer file.Close()

	zw := zip.NewWriter(file)
	defer zw.Close()

	for fname, licf := range lp.Data {
		entryName := filepath.Base(fname)

		if lp.ZipCompressionLevel > 0 && lp.ZipCompressionLevel <= 9 {
			if err := writeCompressedEntry(zw, entryName, lp.ZipCompressionLevel, &licf); err != nil {
				return err
			}
		} else {
			w, err := zw.Create(entryName)
			if err != nil {
				return fmt.Errorf("creating zip entry %q: %w", entryName, err)
			}
			if err := licf.WriteTo(w, entryName); err != nil {
				return fmt.Errorf("writing %q to zip: %w", entryName, err)
			}
		}
	}

	return nil
}

// writeCompressedEntry — создаёт zip-запись с указанным уровнем deflate.
func writeCompressedEntry(zw *zip.Writer, name string, level int, licf *LicelFile) error {
	// Сериализуем несжатые данные для вычисления CRC и размеров.
	var rawBuf bytes.Buffer
	if err := licf.WriteTo(&rawBuf, name); err != nil {
		return fmt.Errorf("serializing %q: %w", name, err)
	}
	raw := rawBuf.Bytes()

	// Сжимаем в буфер.
	var compBuf bytes.Buffer
	fw, err := flate.NewWriter(&compBuf, level)
	if err != nil {
		return fmt.Errorf("creating deflate writer for %q: %w", name, err)
	}
	if _, err := fw.Write(raw); err != nil {
		fw.Close()
		return fmt.Errorf("compressing %q: %w", name, err)
	}
	if err := fw.Close(); err != nil {
		return fmt.Errorf("closing deflate writer for %q: %w", name, err)
	}
	compressed := compBuf.Bytes()

	h := crc32.NewIEEE()
	if _, err := h.Write(raw); err != nil {
		return fmt.Errorf("calculating crc32 for %q: %w", name, err)
	}

	fh := &zip.FileHeader{
		Name:               name,
		Method:             zip.Deflate,
		CRC32:              h.Sum32(),
		UncompressedSize64: uint64(len(raw)),
		CompressedSize64:   uint64(len(compressed)),
	}
	fh.SetMode(0644)

	w, err := zw.CreateRaw(fh)
	if err != nil {
		return fmt.Errorf("creating raw zip entry %q: %w", name, err)
	}

	if _, err := w.Write(compressed); err != nil {
		return fmt.Errorf("writing compressed data for %q: %w", name, err)
	}

	return nil
}
