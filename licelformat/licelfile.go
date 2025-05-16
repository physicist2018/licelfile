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

	"github.com/archsh/timefmt"
	"github.com/rs/zerolog/log"
)

const (
	LICEL_MAX_HEADER_LEN = 80
)

type LicelProfilesList []LicelProfile

// LicelFile — структура, представляющая единичное измерение
type LicelFile struct {
	MeasurementSite       string            `json:"location"`       // Место измерения
	MeasurementStartTime  time.Time         `json:"start_time"`     // Время начала измерения
	MeasurementStopTime   time.Time         `json:"stop_time"`      // Время окончания измерения
	AltitudeAboveSeaLevel float64           `json:"lidar_altitude"` // Высота над уровнем моря
	Longitude             float64           `json:"longitude"`      // Долгота
	Latitude              float64           `json:"latitude"`       // Широта
	Zenith                float64           `json:"zenith"`         // Зенит
	Laser1NShots          int64             `json:"laser1_nshots"`  // Количество импульсов лазера 1
	Laser1Freq            int64             `json:"laser1_freq"`    // Частота лазера 1
	Laser2NShots          int64             `json:"laser2_nshots"`  // Количество импульсов лазера 2
	Laser2Freq            int64             `json:"laser2_freq"`    // Частота лазера 2
	NDatasets             int64             `json:"dataset_count"`  // Количество наборов данных
	Laser3NShots          int64             `json:"laser3_nshots"`  // Количество импульсов лазера 3
	Laser3Freq            int64             `json:"laser3_freq"`    // Частота лазера 3
	FileLoaded            bool              `json:"-"`              // Файл загружен
	Profiles              LicelProfilesList `json:"datasets"`       // Список профилей
}

// LoadLicelFile — loads LicelFile the specified file name
func LoadLicelFile(fname string) LicelFile {
	f, err := os.Open(fname)
	if err != nil {
		log.Fatal().Err(err).Str("file", fname).Msg("Error opening file")
	}
	defer f.Close()

	r := bufio.NewReader(f)
	var licf LicelFile

	// Пропустить первую строку (обычно пустую или с ненужной информацией)
	readAndTrimLine(r)

	// Вторая строка: базовая информация
	header := readAndTrimLine(r)
	tmp := strings.Fields(header)

	licf.MeasurementSite = tmp[0]
	licf.MeasurementStartTime = parseTime(tmp[1] + " " + tmp[2])
	licf.MeasurementStopTime = parseTime(tmp[3] + " " + tmp[4])
	licf.AltitudeAboveSeaLevel = str2Float(tmp[5])
	licf.Longitude = str2Float(tmp[6])
	licf.Latitude = str2Float(tmp[7])
	licf.Zenith = str2Float(tmp[8])

	// Третья строка: параметры лазеров
	header = readAndTrimLine(r)
	tmp = strings.Fields(header)
	licf.Laser1NShots = str2Int(tmp[0])
	licf.Laser1Freq = str2Int(tmp[1])
	licf.Laser2NShots = str2Int(tmp[2])
	licf.Laser2Freq = str2Int(tmp[3])
	licf.NDatasets = str2Int(tmp[4])
	licf.Laser3NShots = str2Int(tmp[5])
	licf.Laser3Freq = str2Int(tmp[6])

	// Профили
	licf.Profiles = make(LicelProfilesList, licf.NDatasets)
	for i := int64(0); i < licf.NDatasets; i++ {
		header = readAndTrimLine(r)
		licf.Profiles[i] = NewLicelProfile(header)
	}

	// После заголовков — бинарные данные
	skipCRLF(r)

	for i := int64(0); i < licf.NDatasets; i++ {
		prTmp := make([]byte, licf.Profiles[i].NDataPoints*4)
		if _, err := io.ReadFull(r, prTmp); err != nil {
			log.Fatal().Err(err).Msg("Ошибка при чтении бинарных данных")
		}
		licf.Profiles[i].Data = bytesToFloat64Array(prTmp)
		skipCRLF(r)
	}

	licf.FileLoaded = true
	return licf
}

// readAndTrimLine — reads string from reader add thrim to the right
func readAndTrimLine(r *bufio.Reader) string {
	line, err := r.ReadString('\n')
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading string")
	}
	return strings.TrimRight(line, "\t\r ")
}

// skipCRLF — skips CR and LF
func skipCRLF(r *bufio.Reader) {
	crlf := make([]byte, 2)
	if _, err := io.ReadFull(r, crlf); err != nil {
		log.Fatal().Err(err).Msg("Error skipping CRLF")
	}
}

// parseTime — parse datetime string "dd/mm/yyyy hh:mm:ss"
func parseTime(s string) time.Time {
	t, err := timefmt.Strptime(s, "%d/%m/%Y %H:%M:%S")
	if err != nil {
		log.Fatal().Err(err).Msg("Ошибка при парсинге времени")
	}
	return t
}

// str2Bool — converts string to boolean
func str2Bool(str string) bool {
	v, _ := strconv.ParseBool(str)
	return v
}

// str2Int — converts string to int
func str2Int(str string) int64 {
	v, _ := strconv.ParseInt(str, 10, 64)
	return v
}

// str2Float — converts string to float
func str2Float(str string) float64 {
	v, _ := strconv.ParseFloat(str, 64)
	return v
}

// bytesToFloat64Array — converts []byte to []float64
func bytesToFloat64Array(b []byte) []float64 {
	n := len(b) / 4
	arr := make([]float64, n)
	for i := 0; i < n; i++ {
		arr[i] = float64(int32(binary.LittleEndian.Uint32(b[i*4 : (i+1)*4])))
	}
	return arr
}

// SelectCertainWavelength1 — selects certain profile by its wavelength and type from a single file
func (lf *LicelFile) SelectCertainWavelength(isPhoton bool, wavelength float64) LicelProfile {
	for _, v := range lf.Profiles {
		if v.Photon == isPhoton && v.Wavelength == wavelength {
			return v
		}
	}
	return LicelProfile{}
}

// Save - saves licel file to disk
func (lf *LicelFile) Save(fname string) error {
	file, err := os.Create(fname + "1")
	if err != nil {
		return err
	}
	defer file.Close()
	file.WriteString(lf.FormatFirstLine(fname))
	file.WriteString(lf.FormatSecondLine())
	file.WriteString(lf.FormatThirdLine())
	for _, i := range lf.Profiles {
		file.WriteString(i.Metadata())
	}
	file.WriteString("\r\n")
	for _, i := range lf.Profiles {
		file.WriteString(i.Profile())
	}
	return nil
}

// FormatFirstLine - returns string with first line of LICEL file
func (lf *LicelFile) FormatFirstLine(fname string) string {
	return fmt.Sprintf(" %-77s\r\n", fname)
}

// FormatSecondLine - returns string with second line of LICEL file
func (lf *LicelFile) FormatSecondLine() string {
	s := fmt.Sprintf(" %s %s %s %s %s %04.0f %06.1f %06.1f %02.0f", lf.MeasurementSite, lf.MeasurementStartTime.Format("02/01/2006"),
		lf.MeasurementStartTime.Format("15:04:05"), lf.MeasurementStopTime.Format("02/01/2006"),
		lf.MeasurementStopTime.Format("15:04:05"), lf.AltitudeAboveSeaLevel, lf.Longitude, lf.Latitude, lf.Zenith)
	return fmt.Sprintf("%-78s\r\n", s)
}

// FormatThirdLine - returns string with third line of LICEL file
func (lf *LicelFile) FormatThirdLine() string {
	s := fmt.Sprintf(" %07d %04d %07d %04d %02d %07d %04d", lf.Laser1NShots, lf.Laser1Freq, lf.Laser2NShots, lf.Laser2Freq,
		lf.NDatasets, lf.Laser3NShots, lf.Laser3Freq)

	return fmt.Sprintf("%-78s\r\n", s)
}

// LoadLicelFileFromReader — loads licel file from reader
func LoadLicelFileFromReader(f io.Reader, size int64) LicelFile {
	// f, err := os.Open(fname)
	// if err != nil {
	// 	log.Fatal().Err(err).Str("file", fname).Msg("Error opening file")
	// }
	// defer f.Close()

	r := bufio.NewReader(f)
	var licf LicelFile

	// Пропустить первую строку (обычно пустую или с ненужной информацией)
	readAndTrimLine(r)

	// Вторая строка: базовая информация
	header := readAndTrimLine(r)
	tmp := strings.Fields(header)

	licf.MeasurementSite = tmp[0]
	licf.MeasurementStartTime = parseTime(tmp[1] + " " + tmp[2])
	licf.MeasurementStopTime = parseTime(tmp[3] + " " + tmp[4])
	licf.AltitudeAboveSeaLevel = str2Float(tmp[5])
	licf.Longitude = str2Float(tmp[6])
	licf.Latitude = str2Float(tmp[7])
	licf.Zenith = str2Float(tmp[8])

	// Третья строка: параметры лазеров
	header = readAndTrimLine(r)
	tmp = strings.Fields(header)
	licf.Laser1NShots = str2Int(tmp[0])
	licf.Laser1Freq = str2Int(tmp[1])
	licf.Laser2NShots = str2Int(tmp[2])
	licf.Laser2Freq = str2Int(tmp[3])
	licf.NDatasets = str2Int(tmp[4])
	licf.Laser3NShots = str2Int(tmp[5])
	licf.Laser3Freq = str2Int(tmp[6])

	// Профили
	licf.Profiles = make(LicelProfilesList, licf.NDatasets)
	for i := int64(0); i < licf.NDatasets; i++ {
		header = readAndTrimLine(r)
		licf.Profiles[i] = NewLicelProfile(header)
	}

	// После заголовков — бинарные данные
	skipCRLF(r)

	for i := int64(0); i < licf.NDatasets; i++ {
		prTmp := make([]byte, licf.Profiles[i].NDataPoints*4)
		if _, err := io.ReadFull(r, prTmp); err != nil {
			log.Fatal().Err(err).Msg("Ошибка при чтении бинарных данных")
		}
		licf.Profiles[i].Data = bytesToFloat64Array(prTmp)
		skipCRLF(r)
	}

	licf.FileLoaded = true
	return licf
}
