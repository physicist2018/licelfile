package licelformat

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLicelPack_SaveToNetCDF3_Roundtrip(t *testing.T) {
	now := time.Date(2024, 6, 10, 12, 0, 0, 0, time.UTC)
	start := now
	stop := now.Add(2 * time.Hour)

	// Создаём тестовый пак с 2 файлами
	pack := &LicelPack{
		StartTime: start,
		StopTime:  stop,
		Data: map[string]LicelFile{
			"/data/file1.dat": {
				MeasurementSite:       "SiteA",
				MeasurementStartTime:  start,
				MeasurementStopTime:   start.Add(30 * time.Minute),
				AltitudeAboveSeaLevel: 20,
				Longitude:             131.9,
				Latitude:              43.1,
				Zenith:                50,
				Laser1NShots:          2001,
				Laser1Freq:            20,
				Laser2NShots:          1000,
				Laser2Freq:            10,
				Laser3NShots:          500,
				Laser3Freq:            5,
				NDatasets:             2,
				Profiles: LicelProfilesList{
					{
						Active:       true,
						Photon:       false,
						LaserType:    1,
						NDataPoints:  3,
						Reserved:     [3]int{1, 0, 0},
						HighVoltage:  0,
						BinWidth:     7.5,
						Wavelength:   355,
						Polarization: "o",
						BinShift:     0,
						DecBinShift:  0,
						AdcBits:      12,
						NShots:       2001,
						DiscrLevel:   0.5,
						DeviceID:     "BT",
						NCrate:       0,
						Data:         []float64{1.1, 2.2, 3.3},
					},
					{
						Active:       true,
						Photon:       true,
						LaserType:    1,
						NDataPoints:  3,
						Reserved:     [3]int{1, 0, 0},
						HighVoltage:  0,
						BinWidth:     7.5,
						Wavelength:   355,
						Polarization: "o",
						BinShift:     0,
						DecBinShift:  0,
						AdcBits:      0,
						NShots:       2000,
						DiscrLevel:   3.1746,
						DeviceID:     "BC",
						NCrate:       0,
						Data:         []float64{100, 200, 300},
					},
				},
			},
			"/data/file2.dat": {
				MeasurementSite:       "SiteB",
				MeasurementStartTime:  start.Add(1 * time.Hour),
				MeasurementStopTime:   start.Add(2 * time.Hour),
				AltitudeAboveSeaLevel: 100,
				Longitude:             132.0,
				Latitude:              44.0,
				Zenith:                45,
				Laser1NShots:          1000,
				Laser1Freq:            10,
				Laser2NShots:          500,
				Laser2Freq:            5,
				Laser3NShots:          0,
				Laser3Freq:            0,
				NDatasets:             1,
				Profiles: LicelProfilesList{
					{
						Active:       true,
						Photon:       false,
						LaserType:    1,
						NDataPoints:  2,
						Reserved:     [3]int{0, 0, 0},
						HighVoltage:  1000,
						BinWidth:     15.0,
						Wavelength:   532,
						Polarization: "p",
						BinShift:     1,
						DecBinShift:  0,
						AdcBits:      14,
						NShots:       500,
						DiscrLevel:   1.0,
						DeviceID:     "BT",
						NCrate:       1,
						Data:         []float64{10.5, 20.7},
					},
				},
			},
		},
	}

	// Сохраняем в NetCDF3
	tmpDir := t.TempDir()
	ncPath := filepath.Join(tmpDir, "pack.nc")

	err := pack.SaveToNetCDF3(ncPath)
	require.NoError(t, err, "SaveToNetCDF3 should succeed")

	// Загружаем обратно
	loaded, err := LoadLicelPackFromNetCDF3(ncPath)
	require.NoError(t, err, "LoadLicelPackFromNetCDF3 should succeed")

	// --- Проверка глобальных полей пакета ---
	assert.True(t, loaded.StartTime.Equal(pack.StartTime),
		"StartTime mismatch: expected %v, got %v", pack.StartTime, loaded.StartTime)
	assert.True(t, loaded.StopTime.Equal(pack.StopTime),
		"StopTime mismatch: expected %v, got %v", pack.StopTime, loaded.StopTime)

	require.Len(t, loaded.Data, len(pack.Data), "number of files should match")

	// --- Проверка каждого файла ---
	for fname, expectedFile := range pack.Data {
		gotFile, ok := loaded.Data[fname]
		require.True(t, ok, "file %q should exist in loaded pack", fname)

		assert.Equal(t, expectedFile.MeasurementSite, gotFile.MeasurementSite, "file %q: site", fname)
		assert.True(t, gotFile.MeasurementStartTime.Equal(expectedFile.MeasurementStartTime),
			"file %q: start time", fname)
		assert.True(t, gotFile.MeasurementStopTime.Equal(expectedFile.MeasurementStopTime),
			"file %q: stop time", fname)
		assert.Equal(t, expectedFile.AltitudeAboveSeaLevel, gotFile.AltitudeAboveSeaLevel,
			"file %q: altitude", fname)
		assert.Equal(t, expectedFile.Longitude, gotFile.Longitude, "file %q: longitude", fname)
		assert.Equal(t, expectedFile.Latitude, gotFile.Latitude, "file %q: latitude", fname)
		assert.Equal(t, expectedFile.Zenith, gotFile.Zenith, "file %q: zenith", fname)
		assert.Equal(t, expectedFile.Laser1NShots, gotFile.Laser1NShots, "file %q: l1_nshots", fname)
		assert.Equal(t, expectedFile.Laser1Freq, gotFile.Laser1Freq, "file %q: l1_freq", fname)
		assert.Equal(t, expectedFile.Laser2NShots, gotFile.Laser2NShots, "file %q: l2_nshots", fname)
		assert.Equal(t, expectedFile.Laser2Freq, gotFile.Laser2Freq, "file %q: l2_freq", fname)
		assert.Equal(t, expectedFile.Laser3NShots, gotFile.Laser3NShots, "file %q: l3_nshots", fname)
		assert.Equal(t, expectedFile.Laser3Freq, gotFile.Laser3Freq, "file %q: l3_freq", fname)
		assert.Equal(t, expectedFile.NDatasets, gotFile.NDatasets, "file %q: ndatasets", fname)
		assert.True(t, gotFile.FileLoaded, "file %q: FileLoaded should be true", fname)

		// Профили
		require.Len(t, gotFile.Profiles, len(expectedFile.Profiles),
			"file %q: number of profiles", fname)

		for pi := range expectedFile.Profiles {
			ep := expectedFile.Profiles[pi]
			gp := gotFile.Profiles[pi]

			assert.Equal(t, ep.Active, gp.Active, "file %q profile %d: Active", fname, pi)
			assert.Equal(t, ep.Photon, gp.Photon, "file %q profile %d: Photon", fname, pi)
			assert.Equal(t, ep.LaserType, gp.LaserType, "file %q profile %d: LaserType", fname, pi)
			assert.Equal(t, ep.NDataPoints, gp.NDataPoints, "file %q profile %d: NDataPoints", fname, pi)
			assert.Equal(t, ep.Reserved, gp.Reserved, "file %q profile %d: Reserved", fname, pi)
			assert.Equal(t, ep.HighVoltage, gp.HighVoltage, "file %q profile %d: HighVoltage", fname, pi)
			assert.Equal(t, ep.BinWidth, gp.BinWidth, "file %q profile %d: BinWidth", fname, pi)
			assert.Equal(t, ep.Wavelength, gp.Wavelength, "file %q profile %d: Wavelength", fname, pi)
			assert.Equal(t, ep.Polarization, gp.Polarization, "file %q profile %d: Polarization", fname, pi)
			assert.Equal(t, ep.BinShift, gp.BinShift, "file %q profile %d: BinShift", fname, pi)
			assert.Equal(t, ep.DecBinShift, gp.DecBinShift, "file %q profile %d: DecBinShift", fname, pi)
			assert.Equal(t, ep.AdcBits, gp.AdcBits, "file %q profile %d: AdcBits", fname, pi)
			assert.Equal(t, ep.NShots, gp.NShots, "file %q profile %d: NShots", fname, pi)
			assert.InDelta(t, ep.DiscrLevel, gp.DiscrLevel, 1e-6,
				"file %q profile %d: DiscrLevel", fname, pi)
			assert.Equal(t, ep.DeviceID, gp.DeviceID, "file %q profile %d: DeviceID", fname, pi)
			assert.Equal(t, ep.NCrate, gp.NCrate, "file %q profile %d: NCrate", fname, pi)

			require.Len(t, gp.Data, len(ep.Data),
				"file %q profile %d: data length", fname, pi)
			for di := range ep.Data {
				assert.InDelta(t, ep.Data[di], gp.Data[di], 1e-9,
					"file %q profile %d data[%d]", fname, pi, di)
			}
		}
	}
}

func TestLicelPack_SaveToNetCDF3_Empty(t *testing.T) {
	pack := &LicelPack{
		StartTime: time.Time{},
		StopTime:  time.Time{},
		Data:      map[string]LicelFile{},
	}

	tmpDir := t.TempDir()
	ncPath := filepath.Join(tmpDir, "empty.nc")

	err := pack.SaveToNetCDF3(ncPath)
	assert.Error(t, err, "empty pack should return error")
	assert.Contains(t, err.Error(), "empty")
}

func TestLicelPack_SaveToNetCDF3_FileWithGluedProfile(t *testing.T) {
	now := time.Date(2024, 6, 10, 12, 0, 0, 0, time.UTC)

	pack := &LicelPack{
		StartTime: now,
		StopTime:  now,
		Data: map[string]LicelFile{
			"glued.dat": {
				MeasurementSite:       "Test",
				MeasurementStartTime:  now,
				MeasurementStopTime:   now,
				AltitudeAboveSeaLevel: 10,
				Longitude:             0,
				Latitude:              0,
				Zenith:                0,
				Laser1NShots:          100,
				Laser1Freq:            10,
				NDatasets:             1,
				Profiles: LicelProfilesList{
					{
						Active:       true,
						Photon:       false,
						LaserType:    1,
						NDataPoints:  2,
						Reserved:     [3]int{0, 0, 0},
						HighVoltage:  0,
						BinWidth:     7.5,
						Wavelength:   355,
						Polarization: "o",
						BinShift:     0,
						DecBinShift:  0,
						AdcBits:      12,
						NShots:       100,
						DiscrLevel:   0.5,
						DeviceID:     "BG",
						NCrate:       0,
						Data:         []float64{1000, 2000},
					},
				},
			},
		},
	}

	tmpDir := t.TempDir()
	ncPath := filepath.Join(tmpDir, "glued.nc")

	err := pack.SaveToNetCDF3(ncPath)
	require.NoError(t, err)

	loaded, err := LoadLicelPackFromNetCDF3(ncPath)
	require.NoError(t, err)

	require.Len(t, loaded.Data, 1)
	glued, ok := loaded.Data["glued.dat"]
	require.True(t, ok)
	require.Len(t, glued.Profiles, 1)
	assert.Equal(t, "BG", glued.Profiles[0].DeviceID)
	assert.True(t, glued.Profiles[0].IsGlued())
	assert.Equal(t, []float64{1000, 2000}, glued.Profiles[0].Data)
}
