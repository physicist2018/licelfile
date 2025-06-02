package licelformat

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
)

const (
	LICEL_MAX_RESERVED = 3
)

// LicelProfile — структура, представляющая измерительный канал
type LicelProfile struct {
	Active       bool                    `json:"is_active"`
	Photon       bool                    `json:"is_photon"`     // Активность канала и тип измерения (фотоны или нет)
	LaserType    int                     `json:"laser_type"`    // Тип лазера
	NDataPoints  int                     `json:"data_points"`   // Количество данных
	Reserved     [LICEL_MAX_RESERVED]int `json:"reserved"`      // Резервные значения
	HighVoltage  int                     `json:"high_voltage"`  // Напряжение
	BinWidth     float64                 `json:"bin_width"`     // Ширина бина
	Wavelength   float64                 `json:"wavelength"`    // Длина волны
	Polarization string                  `json:"polarization"`  // Поляризация
	BinShift     int                     `json:"bin_shift"`     // Сдвиг бина
	DecBinShift  int                     `json:"dec_bin_shift"` // Децибельный сдвиг
	AdcBits      int                     `json:"adc_bits"`      // Биты АЦП
	NShots       int                     `json:"n_shots"`       // Количество импульсов
	DiscrLevel   float64                 `json:"discr_level"`   // Уровень дискриминации
	DeviceID     string                  `json:"device_id"`     // Идентификатор устройства
	NCrate       int                     `json:"n_crate"`       // Номер устройства в крэйте
	Data         []float64               `json:"data"`          // Данные
}

// NewLicelProfile — parse string line into LicelProfile
func NewLicelProfile(line string) LicelProfile {
	items := strings.Fields(line)
	wvlpol := strings.SplitN(items[7], ".", 2)

	return LicelProfile{
		Active:       str2Bool(items[0]),
		Photon:       str2Bool(items[1]),
		LaserType:    str2Int(items[2]),
		NDataPoints:  str2Int(items[3]),
		Reserved:     [3]int{str2Int(items[4]), str2Int(items[8]), str2Int(items[9])},
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

// Metadata — возвращает метаданные канала
func (lp *LicelProfile) Metadata() string {
	var s string
	if lp.Photon {
		s = fmt.Sprintf(" %1d %1d %1d %05d %1d %04d %04.2f %05d.%1s %0d %0d %02d %03d %02d %06d %05.4f %2s%01d", btoi(lp.Active), btoi(lp.Photon), lp.LaserType, lp.NDataPoints,
			lp.Reserved[0], lp.HighVoltage, lp.BinWidth, int(lp.Wavelength), lp.Polarization, 0, 0, lp.BinShift, lp.DecBinShift,
			lp.AdcBits, lp.NShots, lp.DiscrLevel, lp.DeviceID, lp.NCrate)
	} else {
		s = fmt.Sprintf(" %1d %1d %1d %05d %1d %04d %04.2f %05d.%1s %0d %0d %02d %03d %02d %06d %05.3f %2s%01d", btoi(lp.Active), btoi(lp.Photon), lp.LaserType, lp.NDataPoints,
			lp.Reserved[0], lp.HighVoltage, lp.BinWidth, int(lp.Wavelength), lp.Polarization, 0, 0, lp.BinShift, lp.DecBinShift,
			lp.AdcBits, lp.NShots, lp.DiscrLevel, lp.DeviceID, lp.NCrate)
	}
	return fmt.Sprintf("%-78s\r\n", s)
}

// Profile — преобразование данных канала в строку
func (lp *LicelProfile) Profile() string {
	r, err := float64toInt32Bytes(lp.Data)
	if err != nil {
		return "\r\n"
	} else {
		return string(r) + "\r\n"
	}
}

// float64toInt32Bytes — преобразование массива вещественных чисел в массив байтов
func float64toInt32Bytes(data []float64) ([]byte, error) {
	buf := new(bytes.Buffer)
	for _, num := range data {
		err := binary.Write(buf, binary.LittleEndian, int32(num))
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// str2Bool — преобразование строки в логический тип
func btoi(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}
