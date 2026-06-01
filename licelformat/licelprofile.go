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

// LicelProfilesList — список профилей
type LicelProfilesList []LicelProfile

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
func NewLicelProfile(line string) (LicelProfile, error) {
	items := strings.Fields(line)
	if len(items) < 16 {
		return LicelProfile{}, fmt.Errorf("profile header: expected at least 16 fields, got %d", len(items))
	}

	wvlpol := strings.SplitN(items[7], ".", 2)
	if len(wvlpol) != 2 {
		return LicelProfile{}, fmt.Errorf("profile header: invalid wavelength.polarization format %q", items[7])
	}

	wavelength, err := str2Float(wvlpol[0])
	if err != nil {
		return LicelProfile{}, fmt.Errorf("parsing wavelength %q: %w", wvlpol[0], err)
	}

	active, err := str2Bool(items[0])
	if err != nil {
		return LicelProfile{}, fmt.Errorf("parsing active %q: %w", items[0], err)
	}
	photon, err := str2Bool(items[1])
	if err != nil {
		return LicelProfile{}, fmt.Errorf("parsing photon %q: %w", items[1], err)
	}
	laserType, err := str2Int(items[2])
	if err != nil {
		return LicelProfile{}, fmt.Errorf("parsing laser type %q: %w", items[2], err)
	}
	nDataPoints, err := str2Int(items[3])
	if err != nil {
		return LicelProfile{}, fmt.Errorf("parsing data points %q: %w", items[3], err)
	}
	reserved0, err := str2Int(items[4])
	if err != nil {
		return LicelProfile{}, fmt.Errorf("parsing reserved[0] %q: %w", items[4], err)
	}
	highVoltage, err := str2Int(items[5])
	if err != nil {
		return LicelProfile{}, fmt.Errorf("parsing high voltage %q: %w", items[5], err)
	}
	binWidth, err := str2Float(items[6])
	if err != nil {
		return LicelProfile{}, fmt.Errorf("parsing bin width %q: %w", items[6], err)
	}
	reserved1, err := str2Int(items[8])
	if err != nil {
		return LicelProfile{}, fmt.Errorf("parsing reserved[1] %q: %w", items[8], err)
	}
	reserved2, err := str2Int(items[9])
	if err != nil {
		return LicelProfile{}, fmt.Errorf("parsing reserved[2] %q: %w", items[9], err)
	}
	binShift, err := str2Int(items[10])
	if err != nil {
		return LicelProfile{}, fmt.Errorf("parsing bin shift %q: %w", items[10], err)
	}
	decBinShift, err := str2Int(items[11])
	if err != nil {
		return LicelProfile{}, fmt.Errorf("parsing dec bin shift %q: %w", items[11], err)
	}
	adcBits, err := str2Int(items[12])
	if err != nil {
		return LicelProfile{}, fmt.Errorf("parsing adc bits %q: %w", items[12], err)
	}
	nShots, err := str2Int(items[13])
	if err != nil {
		return LicelProfile{}, fmt.Errorf("parsing n shots %q: %w", items[13], err)
	}
	discrLevel, err := str2Float(items[14])
	if err != nil {
		return LicelProfile{}, fmt.Errorf("parsing discr level %q: %w", items[14], err)
	}

	if len(items[15]) < 3 {
		return LicelProfile{}, fmt.Errorf("profile header: device field %q too short", items[15])
	}
	deviceID := items[15][:2]
	nCrate, err := str2Int(items[15][2:])
	if err != nil {
		return LicelProfile{}, fmt.Errorf("parsing n crate %q: %w", items[15][2:], err)
	}

	return LicelProfile{
		Active:       active,
		Photon:       photon,
		LaserType:    laserType,
		NDataPoints:  nDataPoints,
		Reserved:     [3]int{reserved0, reserved1, reserved2},
		HighVoltage:  highVoltage,
		BinWidth:     binWidth,
		Wavelength:   wavelength,
		Polarization: wvlpol[1],
		BinShift:     binShift,
		DecBinShift:  decBinShift,
		AdcBits:      adcBits,
		NShots:       nShots,
		DiscrLevel:   discrLevel,
		DeviceID:     deviceID,
		NCrate:       nCrate,
	}, nil
}

// Metadata — возвращает строку с метаданными канала для записи в заголовок файла
func (lp *LicelProfile) Metadata() string {
	var s string
	if lp.Photon {
		s = fmt.Sprintf(" %1d %1d %1d %05d %1d %04d %04.2f %05d.%1s %0d %0d %02d %03d %02d %06d %05.4f %2s%01d",
			btoi(lp.Active), btoi(lp.Photon), lp.LaserType, lp.NDataPoints,
			lp.Reserved[0], lp.HighVoltage, lp.BinWidth, int(lp.Wavelength), lp.Polarization,
			lp.Reserved[1], lp.Reserved[2], lp.BinShift, lp.DecBinShift,
			lp.AdcBits, lp.NShots, lp.DiscrLevel, lp.DeviceID, lp.NCrate)
	} else {
		s = fmt.Sprintf(" %1d %1d %1d %05d %1d %04d %04.2f %05d.%1s %0d %0d %02d %03d %02d %06d %05.3f %2s%01d",
			btoi(lp.Active), btoi(lp.Photon), lp.LaserType, lp.NDataPoints,
			lp.Reserved[0], lp.HighVoltage, lp.BinWidth, int(lp.Wavelength), lp.Polarization,
			lp.Reserved[1], lp.Reserved[2], lp.BinShift, lp.DecBinShift,
			lp.AdcBits, lp.NShots, lp.DiscrLevel, lp.DeviceID, lp.NCrate)
	}
	return fmt.Sprintf("%-78s\r\n", s)
}

// scaleFactor вычисляет масштабирующий коэффициент для данных профиля
func (lp *LicelProfile) scaleFactor() float64 {
	if lp.Photon {
		return 1.0 / (float64(lp.NShots) * 0.05)
	}
	adcScale := 1 << lp.AdcBits
	return lp.DiscrLevel * 1000.0 / float64(adcScale*lp.NShots)
}

// ProfileRaw — возвращает unscaled бинарное представление данных канала
func (lp *LicelProfile) ProfileRaw() ([]byte, error) {
	scale := lp.scaleFactor()
	unscaled := make([]float64, len(lp.Data))
	for i, v := range lp.Data {
		unscaled[i] = v / scale
	}
	return float64toInt32Bytes(unscaled)
}

// float64toInt32Bytes — преобразование массива float64 в []byte (little-endian int32)
func float64toInt32Bytes(data []float64) ([]byte, error) {
	buf := new(bytes.Buffer)
	for _, num := range data {
		if err := binary.Write(buf, binary.LittleEndian, int32(num)); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// SetMaxDist обрезает данные профиля до дальности alt (метры).
// idx = alt / BinWidth. Ошибка если idx ≤ 0 или idx > NDataPoints.
func (lp *LicelProfile) SetMaxDist(alt float64) error {
	if lp.BinWidth <= 0 {
		return fmt.Errorf("SetMaxDist: bin width must be positive, got %.2f", lp.BinWidth)
	}
	idx := int(alt / lp.BinWidth)
	if idx <= 0 {
		return fmt.Errorf("SetMaxDist: alt %.0f m → idx %d, must be > 0", alt, idx)
	}
	if idx > lp.NDataPoints {
		return fmt.Errorf("SetMaxDist: alt %.0f m → idx %d exceeds NDataPoints %d", alt, idx, lp.NDataPoints)
	}
	lp.Data = lp.Data[:idx]
	lp.NDataPoints = len(lp.Data)
	return nil
}

// btoi — bool to int (1/0)
func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}
