package licelformat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewLicelProfile(t *testing.T) {
	line := " 1 0 1 16380 1 0000 7.50 00355.o 0 0 00 000 12 002001 0.500 BT0"
	profile := NewLicelProfile(line)

	// Проверка значений полей профиля
	assert.Equal(t, true, profile.Active)
	assert.Equal(t, false, profile.Photon)
	assert.Equal(t, int64(1), profile.LaserType)
	assert.Equal(t, int64(16380), profile.NDataPoints)
	assert.Equal(t, int64(1), profile.Reserved[0])
	assert.Equal(t, int64(0), profile.Reserved[1])
	assert.Equal(t, int64(0), profile.Reserved[2])
	assert.Equal(t, int64(0000), profile.HighVoltage)
	assert.Equal(t, 7.5, profile.BinWidth)
	assert.Equal(t, 355.0, profile.Wavelength)
	assert.Equal(t, "o", profile.Polarization)
	assert.Equal(t, int64(0), profile.BinShift)
	assert.Equal(t, int64(2001), profile.NShots)
	assert.Equal(t, 0.5, profile.DiscrLevel)
	assert.Equal(t, "BT", profile.DeviceID)
	assert.Equal(t, int64(0), profile.NCrate)
}

func TestLoadLicelFile(t *testing.T) {
	// Создаем временный файл для теста
	testFile := "../testdata/b2021019.223500"

	// Загружаем файл
	licelFile := LoadLicelFile(testFile)

	// Проверка значений из файла
	assert.True(t, licelFile.FileLoaded)
	assert.Equal(t, "Vladivos", licelFile.MeasurementSite)
	assert.Equal(t, time.Date(2020, 2, 10, 19, 22, 35, 0, time.UTC), licelFile.MeasurementStartTime)
	assert.Equal(t, time.Date(2020, 2, 10, 19, 24, 15, 0, time.UTC), licelFile.MeasurementStopTime)
	assert.Equal(t, float64(20), licelFile.AltitudeAboveSeaLevel)
	assert.Equal(t, float64(50), licelFile.Zenith)
	assert.Equal(t, 131.9, licelFile.Longitude)
	assert.Equal(t, 43.1, licelFile.Latitude)
}

func TestSelectCertainWavelength1(t *testing.T) {
	// Подготавливаем тестовые данные
	profile := LicelProfile{
		Active:     true,
		Photon:     true,
		Wavelength: 400.0,
	}
	licelFile := LicelFile{
		Profiles: []LicelProfile{profile},
	}

	// Выбираем профиль по длине волны
	selectedProfile := SelectCertainWavelength1(&licelFile, true, 400.0)

	// Проверка результата
	assert.Equal(t, profile, selectedProfile)
}

func TestSelectCertainWavelength2(t *testing.T) {
	// Подготавливаем тестовые данные
	profile1 := LicelProfile{
		Active:     true,
		Photon:     true,
		Wavelength: 400.0,
	}
	profile2 := LicelProfile{
		Active:     false,
		Photon:     true,
		Wavelength: 500.0,
	}
	licelPack := LicelPack{
		"file1": LicelFile{Profiles: []LicelProfile{profile1}},
		"file2": LicelFile{Profiles: []LicelProfile{profile2}},
	}

	// Выбираем все профили по длине волны 400.0
	profiles := SelectCertainWavelength2(&licelPack, true, 400.0)

	// Проверка результата
	assert.Len(t, profiles, 1)
	assert.Equal(t, profile1, profiles[0])
}

func TestStr2Bool(t *testing.T) {
	assert.True(t, str2Bool("true"))
	assert.False(t, str2Bool("false"))
}

func TestStr2Int(t *testing.T) {
	assert.Equal(t, int64(100), str2Int("100"))
	assert.Equal(t, int64(-50), str2Int("-50"))
}

func TestStr2Float(t *testing.T) {
	assert.Equal(t, 10.5, str2Float("10.5"))
	assert.Equal(t, 0.0, str2Float("0"))
}
