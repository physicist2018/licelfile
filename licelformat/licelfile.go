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
		licf.Profiles[i], err = newLicelProfile(header)
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
	return time.ParseInLocation("02/01/2006 15:04:05", s, time.Local)
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

// SelectProfile — returns a profile matching photon flag, wavelength, and polarization.
// Pass "" for polarization to match any.
// Returns (LicelProfile{}, false) if no match found.
func (lf *LicelFile) SelectProfile(isPhoton bool, wavelength float64, polarization string) (LicelProfile, bool) {
	for _, v := range lf.Profiles {
		if v.IsGlued() {
			continue
		}
		if v.IsPhoton() == isPhoton && v.Wavelength == wavelength {
			if polarization == "" || v.Polarization == polarization {
				return v, true
			}
		}
	}
	return LicelProfile{}, false
}

// Glue склеивает аналоговый и цифровой каналы для заданной длины волны.
//
// Параметры:
//   - wvl — длина волны
//   - h1, h2 — диапазон высот в метрах для вычисления коэффициента склейки
//
// Алгоритм:
//  1. Находит аналоговый (Photon=false) и цифровой (Photon=true) профили.
//  2. В диапазоне [h1; h2] вычисляет среднее отношение analog/photon — коэффициент k.
//  3. Создаёт новый профиль с DeviceID="BG":
//     - h < h1: данные аналогового канала
//     - h1 ≤ h ≤ h2: 0.5*(analog + k*photon)
//     - h > h2: k*photon
func (lf *LicelFile) Glue(wvl float64, h1, h2 float64, polarization string) (LicelProfile, error) {
	analog, ok := lf.SelectProfile(false, wvl, polarization)
	if !ok {
		return LicelProfile{}, fmt.Errorf("glue: analog channel not found for wavelength %.0f", wvl)
	}
	photon, ok := lf.SelectProfile(true, wvl, polarization)
	if !ok {
		return LicelProfile{}, fmt.Errorf("glue: photon channel not found for wavelength %.0f", wvl)
	}

	if analog.BinWidth <= 0 {
		return LicelProfile{}, fmt.Errorf("glue: analog channel has invalid bin width %.2f", analog.BinWidth)
	}
	if photon.BinWidth <= 0 {
		return LicelProfile{}, fmt.Errorf("glue: photon channel has invalid bin width %.2f", photon.BinWidth)
	}

	if h1 >= h2 {
		return LicelProfile{}, fmt.Errorf("glue: h1 (%.2f) must be less than h2 (%.2f)", h1, h2)
	}

	idx1 := int(h1 / analog.BinWidth)
	idx2 := int(h2 / analog.BinWidth)

	dataLen := len(analog.Data)
	if len(photon.Data) < dataLen {
		dataLen = len(photon.Data)
	}

	if idx1 < 0 || idx1 >= dataLen {
		return LicelProfile{}, fmt.Errorf("glue: h1 (%.2f m) maps to index %d, out of range [0, %d)", h1, idx1, dataLen)
	}
	if idx2 >= dataLen {
		return LicelProfile{}, fmt.Errorf("glue: h2 (%.2f m) maps to index %d, exceeds data length %d", h2, idx2, dataLen)
	}

	// Вычисляем k как среднее отношение analog/photon на [idx1:idx2]
	n := 0
	var sumK float64
	for i := idx1; i <= idx2; i++ {
		if photon.Data[i] == 0 {
			continue
		}
		sumK += analog.Data[i] / photon.Data[i]
		n++
	}
	if n == 0 {
		return LicelProfile{}, fmt.Errorf("glue: all photon data values are zero in range [%.2f, %.2f], cannot compute coefficient", h1, h2)
	}
	k := sumK / float64(n)

	// Создаём результирующий профиль
	result := LicelProfile{
		Active:       true,
		Photon:       false,
		LaserType:    analog.LaserType,
		NDataPoints:  dataLen,
		Reserved:     analog.Reserved,
		HighVoltage:  analog.HighVoltage,
		BinWidth:     analog.BinWidth,
		Wavelength:   analog.Wavelength,
		Polarization: analog.Polarization,
		BinShift:     analog.BinShift,
		DecBinShift:  analog.DecBinShift,
		AdcBits:      analog.AdcBits,
		NShots:       analog.NShots,
		DiscrLevel:   analog.DiscrLevel,
		DeviceID:     "BG",
		NCrate:       analog.NCrate,
		Data:         make([]float64, dataLen),
	}

	for i := 0; i < dataLen; i++ {
		switch {
		case i < idx1:
			result.Data[i] = analog.Data[i]
		case i <= idx2:
			result.Data[i] = 0.5 * (analog.Data[i] + k*photon.Data[i])
		default:
			result.Data[i] = k * photon.Data[i]
		}
	}

	return result, nil
}

// SetMaxDist обрезает все профили до дальности alt (метры).
func (lf *LicelFile) SetMaxDist(alt float64) error {
	for i := range lf.Profiles {
		if err := lf.Profiles[i].SetMaxDist(alt); err != nil {
			return fmt.Errorf("profile %d: %w", i, err)
		}
	}
	return nil
}

// WriteTo — сериализует LICEL-файл в io.Writer
func (lf *LicelFile) WriteTo(w io.Writer, fname string) error {
	bw := bufio.NewWriter(w)

	if _, err := bw.WriteString(lf.formatFirstLine(fname)); err != nil {
		return fmt.Errorf("writing line 1: %w", err)
	}
	if _, err := bw.WriteString(lf.formatSecondLine()); err != nil {
		return fmt.Errorf("writing line 2: %w", err)
	}
	if _, err := bw.WriteString(lf.formatThirdLine()); err != nil {
		return fmt.Errorf("writing line 3: %w", err)
	}
	for i, p := range lf.Profiles {
		if _, err := bw.WriteString(p.metadata()); err != nil {
			return fmt.Errorf("writing metadata for profile %d: %w", i, err)
		}
	}
	if _, err := bw.WriteString("\r\n"); err != nil {
		return fmt.Errorf("writing header/body separator: %w", err)
	}
	for i, p := range lf.Profiles {
		data, err := p.profileRaw()
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
func (lf *LicelFile) formatFirstLine(fname string) string {
	return fmt.Sprintf(" %-77s\r\n", fname)
}

// FormatSecondLine — форматирует вторую строку LICEL-файла (метаданные измерения)
func (lf *LicelFile) formatSecondLine() string {
	s := fmt.Sprintf(" %s %s %s %s %s %04.0f %06.1f %06.1f %02.0f",
		lf.MeasurementSite,
		lf.MeasurementStartTime.Local().Format("02/01/2006"),
		lf.MeasurementStartTime.Local().Format("15:04:05"),
		lf.MeasurementStopTime.Local().Format("02/01/2006"),
		lf.MeasurementStopTime.Local().Format("15:04:05"),
		lf.AltitudeAboveSeaLevel,
		lf.Longitude,
		lf.Latitude,
		lf.Zenith,
	)
	return fmt.Sprintf("%-78s\r\n", s)
}

// FormatThirdLine — форматирует третью строку LICEL-файла (параметры лазеров)
func (lf *LicelFile) formatThirdLine() string {
	s := fmt.Sprintf(" %07d %04d %07d %04d %02d %07d %04d",
		lf.Laser1NShots, lf.Laser1Freq,
		lf.Laser2NShots, lf.Laser2Freq,
		lf.NDatasets,
		lf.Laser3NShots, lf.Laser3Freq,
	)
	return fmt.Sprintf("%-78s\r\n", s)
}
