package licelformat

import (
	"bufio"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/archsh/timefmt"
)

const (
	LICEL_MAX_HEADER_LEN = 80
	LICEL_MAX_RESERVED   = 3
)

// LicelProfile — структура, представляющая измерительный канал
type LicelProfile struct {
	Active, Photon bool                      // Активность канала и тип измерения (фотоны или нет)
	LaserType      int64                     // Тип лазера
	NDataPoints    int64                     // Количество данных
	Reserved       [LICEL_MAX_RESERVED]int64 // Резервные значения
	HighVoltage    int64                     // Напряжение
	BinWidth       float64                   // Ширина бина
	Wavelength     float64                   // Длина волны
	Polarization   string                    // Поляризация
	BinShift       int64                     // Сдвиг бина
	DecBinShift    int64                     // Децибельный сдвиг
	AdcBits        int64                     // Биты АЦП
	NShots         int64                     // Количество импульсов
	DiscrLevel     float64                   // Уровень дискриминации
	DeviceID       string                    // Идентификатор устройства
	NCrate         int64                     // Частота дискретизации
	Data           []float64                 // Данные
}

type LicelProfilesList []LicelProfile

// LicelFile — структура, представляющая единичное измерение
type LicelFile struct {
	MeasurementSite       string            // Место измерения
	MeasurementStartTime  time.Time         // Время начала измерения
	MeasurementStopTime   time.Time         // Время окончания измерения
	AltitudeAboveSeaLevel float64           // Высота над уровнем моря
	Longitude             float64           // Долгота
	Latitude              float64           // Широта
	Zenith                float64           // Зенит
	Laser1NShots          int64             // Количество импульсов лазера 1
	Laser1Freq            int64             // Частота лазера 1
	Laser2NShots          int64             // Количество импульсов лазера 2
	Laser2Freq            int64             // Частота лазера 2
	NDatasets             int64             // Количество наборов данных
	Laser3NShots          int64             // Количество импульсов лазера 3
	Laser3Freq            int64             // Частота лазера 3
	FileLoaded            bool              // Файл загружен
	Profiles              LicelProfilesList // Список профилей
}

type LicelPack map[string]LicelFile

// NewLicelProfile — parse string line into LicelProfile
func NewLicelProfile(line string) LicelProfile {
	items := strings.Fields(line)
	wvlpol := strings.SplitN(items[7], ".", 2)

	return LicelProfile{
		Active:       str2Bool(items[0]),
		Photon:       str2Bool(items[1]),
		LaserType:    str2Int(items[2]),
		NDataPoints:  str2Int(items[3]),
		Reserved:     [3]int64{str2Int(items[4]), str2Int(items[8]), str2Int(items[9])},
		HighVoltage:  str2Int(items[5]),
		BinWidth:     str2Float(items[6]),
		Wavelength:   str2Float(wvlpol[0]),
		Polarization: wvlpol[1],
		BinShift:     str2Int(items[10]),
		DecBinShift:  str2Int(items[11]),
		AdcBits:      str2Int(items[12]),
		NShots:       str2Int(items[13]),
		DiscrLevel:   str2Float(items[14]),
		DeviceID:     items[15][:2],
		NCrate:       str2Int(items[15][2:]),
	}
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

// NewLicelPack — loads files according to mask
func NewLicelPack(mask string) LicelPack {
	pack := make(LicelPack)
	files, err := filepath.Glob(mask)
	if err != nil {
		log.Fatal().Err(err).Str("mask", mask).Msg("Error getting files by mask")
	}
	for _, fname := range files {
		pack[fname] = LoadLicelFile(fname)
	}
	return pack
}

// SelectCertainWavelength1 — selects certain profile by its wavelength and type from a single file
func SelectCertainWavelength1(lf *LicelFile, isPhoton bool, wavelength float64) LicelProfile {
	for _, v := range lf.Profiles {
		if v.Photon == isPhoton && v.Wavelength == wavelength {
			return v
		}
	}
	return LicelProfile{}
}

// SelectCertainWavelength2 — selects certain profile by its wavelength and type from a LicelPack
func SelectCertainWavelength2(lp *LicelPack, isPhoton bool, wavelength float64) LicelProfilesList {
	var result LicelProfilesList
	for _, file := range *lp {
		profile := SelectCertainWavelength1(&file, isPhoton, wavelength)
		if profile.Wavelength != 0 {
			result = append(result, profile)
		}
	}
	return result
}
