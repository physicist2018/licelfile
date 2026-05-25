package licelformat

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	LICEL_MAX_HEADER_LEN = 80
)

// LicelFile — структура, представляющая единичное измерение
type LicelFile struct {
	MeasurementSite       string            `json:"location"`       // Место измерения
	MeasurementStartTime  time.Time         `json:"start_time"`     // Время начала измерения
	MeasurementStopTime   time.Time         `json:"stop_time"`      // Время окончания измерения
	AltitudeAboveSeaLevel float64           `json:"lidar_altitude"` // Высота над уровнем моря
	Longitude             float64           `json:"longitude"`      // Долгота
	Latitude              float64           `json:"latitude"`       // Широта
	Zenith                float64           `json:"zenith"`         // Зенит
	Laser1NShots          int               `json:"laser1_nshots"`  // Количество импульсов лазера 1
	Laser1Freq            int               `json:"laser1_freq"`    // Частота лазера 1
	Laser2NShots          int               `json:"laser2_nshots"`  // Количество импульсов лазера 2
	Laser2Freq            int               `json:"laser2_freq"`    // Частота лазера 2
	NDatasets             int               `json:"dataset_count"`  // Количество наборов данных
	Laser3NShots          int               `json:"laser3_nshots"`  // Количество импульсов лазера 3
	Laser3Freq            int               `json:"laser3_freq"`    // Частота лазера 3
	FileLoaded            bool              `json:"-"`              // Файл загружен
	Profiles              LicelProfilesList `json:"datasets"`       // Список профилей
}

// LoadLicelFile — загружает LICEL-файл по имени
func LoadLicelFile(fname string) (LicelFile, error) {
	f, err := os.Open(fname)
	if err != nil {
		return LicelFile{}, fmt.Errorf("opening file %q: %w", fname, err)
	}
	defer f.Close()

	return loadFromReader(bufio.NewReader(f))
}

// LoadLicelFileFromReader — загружает LICEL-файл из произвольного io.Reader
func LoadLicelFileFromReader(r io.Reader) (LicelFile, error) {
	return loadFromReader(bufio.NewReader(r))
}

// loadFromReader — общая логика загрузки LICEL-файла
func loadFromReader(r *bufio.Reader) (LicelFile, error) {
	var licf LicelFile

	// Пропустить первую строку (обычно пустую или с именем файла)
	if _, err := readAndTrimLine(r); err != nil {
		return licf, fmt.Errorf("reading line 1: %w", err)
	}

	// Вторая строка: базовая информация
	header, err := readAndTrimLine(r)
	if err != nil {
		return licf, fmt.Errorf("reading line 2: %w", err)
	}
	tmp := strings.Fields(header)
	if len(tmp) < 9 {
		return licf, fmt.Errorf("line 2: expected at least 9 fields, got %d", len(tmp))
	}

	licf.MeasurementSite = tmp[0]

	licf.MeasurementStartTime, err = parseTime(tmp[1] + " " + tmp[2])
	if err != nil {
		return licf, fmt.Errorf("parsing start time: %w", err)
	}
	licf.MeasurementStopTime, err = parseTime(tmp[3] + " " + tmp[4])
	if err != nil {
		return licf, fmt.Errorf("parsing stop time: %w", err)
	}

	var fErr error
	licf.AltitudeAboveSeaLevel, fErr = str2Float(tmp[5])
	if fErr != nil {
		return licf, fmt.Errorf("parsing altitude: %w", fErr)
	}
	licf.Longitude, fErr = str2Float(tmp[6])
	if fErr != nil {
		return licf, fmt.Errorf("parsing longitude: %w", fErr)
	}
	licf.Latitude, fErr = str2Float(tmp[7])
	if fErr != nil {
		return licf, fmt.Errorf("parsing latitude: %w", fErr)
	}
	licf.Zenith, fErr = str2Float(tmp[8])
	if fErr != nil {
		return licf, fmt.Errorf("parsing zenith: %w", fErr)
	}

	// Третья строка: параметры лазеров
	header, err = readAndTrimLine(r)
	if err != nil {
		return licf, fmt.Errorf("reading line 3: %w", err)
	}
	tmp = strings.Fields(header)
	if len(tmp) < 7 {
		return licf, fmt.Errorf("line 3: expected at least 7 fields, got %d", len(tmp))
	}

	var iErr error
	licf.Laser1NShots, iErr = str2Int(tmp[0])
	if iErr != nil {
		return licf, fmt.Errorf("parsing laser1 nshots: %w", iErr)
	}
	licf.Laser1Freq, iErr = str2Int(tmp[1])
	if iErr != nil {
		return licf, fmt.Errorf("parsing laser1 freq: %w", iErr)
	}
	licf.Laser2NShots, iErr = str2Int(tmp[2])
	if iErr != nil {
		return licf, fmt.Errorf("parsing laser2 nshots: %w", iErr)
	}
	licf.Laser2Freq, iErr = str2Int(tmp[3])
	if iErr != nil {
		return licf, fmt.Errorf("parsing laser2 freq: %w", iErr)
	}
	licf.NDatasets, iErr = str2Int(tmp[4])
	if iErr != nil {
		return licf, fmt.Errorf("parsing dataset count: %w", iErr)
	}
	licf.Laser3NShots, iErr = str2Int(tmp[5])
	if iErr != nil {
		return licf, fmt.Errorf("parsing laser3 nshots: %w", iErr)
	}
	licf.Laser3Freq, iErr = str2Int(tmp[6])
	if iErr != nil {
		return licf, fmt.Errorf("parsing laser3 freq: %w", iErr)
	}

	// Профили
	licf.Profiles = make(LicelProfilesList, licf.NDatasets)
	for i := 0; i < licf.NDatasets; i++ {
		header, err = readAndTrimLine(r)
		if err != nil {
			return licf, fmt.Errorf("reading profile header %d: %w", i, err)
		}
		licf.Profiles[i], err = NewLicelProfile(header)
		if err != nil {
			return licf, fmt.Errorf("parsing profile %d: %w", i, err)
		}
	}

	// После заголовков — бинарные данные
	if err := skipCRLF(r); err != nil {
		return licf, fmt.Errorf("skipping header/body separator: %w", err)
	}

	for i := 0; i < licf.NDatasets; i++ {
		prTmp := make([]byte, licf.Profiles[i].NDataPoints*4)
		if _, err := io.ReadFull(r, prTmp); err != nil {
			return licf, fmt.Errorf("reading binary data for profile %d: %w", i, err)
		}
		licf.Profiles[i].Data = bytesToFloat64Array(prTmp)

		scale := licf.Profiles[i].scaleFactor()
		for j := range licf.Profiles[i].Data {
			licf.Profiles[i].Data[j] *= scale
		}
		if err := skipCRLF(r); err != nil {
			return licf, fmt.Errorf("skipping post-profile %d CRLF: %w", i, err)
		}
	}

	licf.FileLoaded = true
	return licf, nil
}

// readAndTrimLine — reads a line from reader and trims whitespace from right
func readAndTrimLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\t\r "), nil
}

// skipCRLF — skips \r\n
func skipCRLF(r *bufio.Reader) error {
	crlf := make([]byte, 2)
	if _, err := io.ReadFull(r, crlf); err != nil {
		return err
	}
	return nil
}

// parseTime — parse datetime string "dd/mm/yyyy hh:mm:ss"
func parseTime(s string) (time.Time, error) {
	return time.ParseInLocation("02/01/2006 15:04:05", s, time.UTC)
}

// str2Bool — converts string to bool
func str2Bool(str string) (bool, error) {
	return strconv.ParseBool(str)
}

// str2Int — converts string to int
func str2Int(str string) (int, error) {
	v, err := strconv.ParseInt(str, 10, 0)
	return int(v), err
}

// str2Float — converts string to float64
func str2Float(str string) (float64, error) {
	return strconv.ParseFloat(str, 64)
}

// bytesToFloat64Array — converts []byte to []float64 (little-endian int32 → float64)
func bytesToFloat64Array(b []byte) []float64 {
	n := len(b) / 4
	arr := make([]float64, n)
	for i := 0; i < n; i++ {
		arr[i] = float64(int32(binary.LittleEndian.Uint32(b[i*4 : (i+1)*4])))
	}
	return arr
}

// SelectProfile — returns a profile matching photon flag and wavelength.
// Returns (LicelProfile{}, false) if no match found.
func (lf *LicelFile) SelectProfile(isPhoton bool, wavelength float64) (LicelProfile, bool) {
	for _, v := range lf.Profiles {
		if v.Photon == isPhoton && v.Wavelength == wavelength {
			return v, true
		}
	}
	return LicelProfile{}, false
}

// WriteTo — сериализует LICEL-файл в io.Writer
func (lf *LicelFile) WriteTo(w io.Writer, fname string) error {
	bw := bufio.NewWriter(w)

	if _, err := bw.WriteString(lf.FormatFirstLine(fname)); err != nil {
		return fmt.Errorf("writing line 1: %w", err)
	}
	if _, err := bw.WriteString(lf.FormatSecondLine()); err != nil {
		return fmt.Errorf("writing line 2: %w", err)
	}
	if _, err := bw.WriteString(lf.FormatThirdLine()); err != nil {
		return fmt.Errorf("writing line 3: %w", err)
	}
	for i, p := range lf.Profiles {
		if _, err := bw.WriteString(p.Metadata()); err != nil {
			return fmt.Errorf("writing metadata for profile %d: %w", i, err)
		}
	}
	if _, err := bw.WriteString("\r\n"); err != nil {
		return fmt.Errorf("writing header/body separator: %w", err)
	}
	for i, p := range lf.Profiles {
		data, err := p.ProfileRaw()
		if err != nil {
			return fmt.Errorf("serializing profile %d: %w", i, err)
		}
		if _, err := bw.Write(data); err != nil {
			return fmt.Errorf("writing binary data for profile %d: %w", i, err)
		}
		if _, err := bw.WriteString("\r\n"); err != nil {
			return fmt.Errorf("writing post-profile %d CRLF: %w", i, err)
		}
	}

	return bw.Flush()
}

// Save — сохраняет LICEL-файл на диск
func (lf *LicelFile) Save(fname string) error {
	file, err := os.Create(fname)
	if err != nil {
		return fmt.Errorf("creating file %q: %w", fname, err)
	}
	defer file.Close()

	return lf.WriteTo(file, fname)
}

// FormatFirstLine — форматирует первую строку LICEL-файла
func (lf *LicelFile) FormatFirstLine(fname string) string {
	return fmt.Sprintf(" %-77s\r\n", fname)
}

// FormatSecondLine — форматирует вторую строку LICEL-файла (метаданные измерения)
func (lf *LicelFile) FormatSecondLine() string {
	s := fmt.Sprintf(" %s %s %s %s %s %04.0f %06.1f %06.1f %02.0f",
		lf.MeasurementSite,
		lf.MeasurementStartTime.UTC().Format("02/01/2006"),
		lf.MeasurementStartTime.UTC().Format("15:04:05"),
		lf.MeasurementStopTime.UTC().Format("02/01/2006"),
		lf.MeasurementStopTime.UTC().Format("15:04:05"),
		lf.AltitudeAboveSeaLevel,
		lf.Longitude,
		lf.Latitude,
		lf.Zenith,
	)
	return fmt.Sprintf("%-78s\r\n", s)
}

// FormatThirdLine — форматирует третью строку LICEL-файла (параметры лазеров)
func (lf *LicelFile) FormatThirdLine() string {
	s := fmt.Sprintf(" %07d %04d %07d %04d %02d %07d %04d",
		lf.Laser1NShots, lf.Laser1Freq,
		lf.Laser2NShots, lf.Laser2Freq,
		lf.NDatasets,
		lf.Laser3NShots, lf.Laser3Freq,
	)
	return fmt.Sprintf("%-78s\r\n", s)
}
