package licelformat

import (
	"fmt"
	"time"

	"github.com/batchatco/go-native-netcdf/netcdf"
	"github.com/batchatco/go-native-netcdf/netcdf/api"
	"github.com/batchatco/go-native-netcdf/netcdf/util"
)

// SaveToNetCDF3 сохраняет LicelPack в NetCDF3 (CDF) файл.
//
// Схема хранения:
//   - Глобальные атрибуты: start_time, stop_time (ISO 8601)
//   - Размерности: nfiles, nprofiles, ndata
//   - Переменные уровня файла (nfiles): filename, measurement_site, …,
//     laser1_nshots, …, ndatasets
//   - Переменные уровня профиля (nprofiles): file_index, active, photon,
//     laser_type, …, device_id, n_crate, data_offset, data_count
//   - profile_data (ndata) — плоский []float64 со всеми данными профилей
func (lp *LicelPack) SaveToNetCDF3(fname string) error {
	nfiles := len(lp.Data)

	if nfiles == 0 {
		// CDF writer не поддерживает пустые размерности; создаём файл только с атрибутами.
		cw, err := netcdf.OpenWriter(fname, netcdf.KindCDF)
		if err != nil {
			return fmt.Errorf("creating netcdf3 file: %w", err)
		}
		defer cw.Close()

		attrs, err := util.NewOrderedMap(
			[]string{"start_time", "stop_time"},
			map[string]any{
				"start_time": lp.StartTime.Format(time.RFC3339),
				"stop_time":  lp.StopTime.Format(time.RFC3339),
			},
		)
		if err != nil {
			return fmt.Errorf("creating global attributes: %w", err)
		}
		return cw.AddAttributes(attrs)
	}

	type flatEntry struct {
		fileIdx   int
		profile   LicelProfile
		dataStart int
	}

	var flat []flatEntry
	var dataBuf []float64
	dataOffset := 0

	fIdx := 0
	for _, lf := range lp.Data {
		for _, pr := range lf.Profiles {
			flat = append(flat, flatEntry{
				fileIdx:   fIdx,
				profile:   pr,
				dataStart: dataOffset,
			})
			dataBuf = append(dataBuf, pr.Data...)
			dataOffset += len(pr.Data)
		}
		fIdx++
	}

	nprofiles := len(flat)

	cw, err := netcdf.OpenWriter(fname, netcdf.KindCDF)
	if err != nil {
		return fmt.Errorf("creating netcdf3 file: %w", err)
	}
	defer cw.Close()

	// --- Global attributes ---
	attrs, err := util.NewOrderedMap(
		[]string{"start_time", "stop_time"},
		map[string]any{
			"start_time": lp.StartTime.Format(time.RFC3339),
			"stop_time":  lp.StopTime.Format(time.RFC3339),
		},
	)
	if err != nil {
		return fmt.Errorf("creating global attributes: %w", err)
	}
	if err := cw.AddAttributes(attrs); err != nil {
		return fmt.Errorf("adding global attributes: %w", err)
	}

	// --- Build file-level slices ---
	filenames := make([]string, nfiles)
	sites := make([]string, nfiles)
	startTimes := make([]string, nfiles)
	stopTimes := make([]string, nfiles)
	altitudes := make([]float64, nfiles)
	longitudes := make([]float64, nfiles)
	latitudes := make([]float64, nfiles)
	zeniths := make([]float64, nfiles)
	l1ns := make([]int32, nfiles)
	l1f := make([]int32, nfiles)
	l2ns := make([]int32, nfiles)
	l2f := make([]int32, nfiles)
	l3ns := make([]int32, nfiles)
	l3f := make([]int32, nfiles)
	ndss := make([]int32, nfiles)

	i := 0
	for fname, lf := range lp.Data {
		filenames[i] = fname
		sites[i] = lf.MeasurementSite
		startTimes[i] = lf.MeasurementStartTime.Format(time.RFC3339)
		stopTimes[i] = lf.MeasurementStopTime.Format(time.RFC3339)
		altitudes[i] = lf.AltitudeAboveSeaLevel
		longitudes[i] = lf.Longitude
		latitudes[i] = lf.Latitude
		zeniths[i] = lf.Zenith
		l1ns[i] = int32(lf.Laser1NShots)
		l1f[i] = int32(lf.Laser1Freq)
		l2ns[i] = int32(lf.Laser2NShots)
		l2f[i] = int32(lf.Laser2Freq)
		l3ns[i] = int32(lf.Laser3NShots)
		l3f[i] = int32(lf.Laser3Freq)
		ndss[i] = int32(lf.NDatasets)
		i++
	}

	// --- File-level variables ---
	dimNfiles := []string{"nfiles"}
	fileVars := []struct {
		name   string
		values any
	}{
		{"filename", filenames},
		{"measurement_site", sites},
		{"measurement_start_time", startTimes},
		{"measurement_stop_time", stopTimes},
		{"altitude", altitudes},
		{"longitude", longitudes},
		{"latitude", latitudes},
		{"zenith", zeniths},
		{"laser1_nshots", l1ns},
		{"laser1_freq", l1f},
		{"laser2_nshots", l2ns},
		{"laser2_freq", l2f},
		{"laser3_nshots", l3ns},
		{"laser3_freq", l3f},
		{"ndatasets", ndss},
	}
	for _, v := range fileVars {
		if err := cw.AddVar(v.name, api.Variable{
			Values:     v.values,
			Dimensions: dimNfiles,
		}); err != nil {
			return fmt.Errorf("adding variable %q: %w", v.name, err)
		}
	}

	// --- Profile-level slices ---
	fileIdxs := make([]int32, nprofiles)
	actives := make([]int32, nprofiles)
	photons := make([]int32, nprofiles)
	laserTypes := make([]int32, nprofiles)
	nDataPoints := make([]int32, nprofiles)
	reserved0 := make([]int32, nprofiles)
	reserved1 := make([]int32, nprofiles)
	reserved2 := make([]int32, nprofiles)
	highVoltages := make([]int32, nprofiles)
	binWidths := make([]float64, nprofiles)
	wavelengths := make([]float64, nprofiles)
	polarizations := make([]string, nprofiles)
	binShifts := make([]int32, nprofiles)
	decBinShifts := make([]int32, nprofiles)
	adcBits := make([]int32, nprofiles)
	nShots := make([]int32, nprofiles)
	discrLevels := make([]float64, nprofiles)
	deviceIDs := make([]string, nprofiles)
	nCrates := make([]int32, nprofiles)
	dataOffsets := make([]int32, nprofiles)
	dataCounts := make([]int32, nprofiles)

	for j, fe := range flat {
		fileIdxs[j] = int32(fe.fileIdx)
		actives[j] = btoi32(fe.profile.Active)
		photons[j] = btoi32(fe.profile.Photon)
		laserTypes[j] = int32(fe.profile.LaserType)
		nDataPoints[j] = int32(fe.profile.NDataPoints)
		reserved0[j] = int32(fe.profile.Reserved[0])
		reserved1[j] = int32(fe.profile.Reserved[1])
		reserved2[j] = int32(fe.profile.Reserved[2])
		highVoltages[j] = int32(fe.profile.HighVoltage)
		binWidths[j] = fe.profile.BinWidth
		wavelengths[j] = fe.profile.Wavelength
		polarizations[j] = fe.profile.Polarization
		binShifts[j] = int32(fe.profile.BinShift)
		decBinShifts[j] = int32(fe.profile.DecBinShift)
		adcBits[j] = int32(fe.profile.AdcBits)
		nShots[j] = int32(fe.profile.NShots)
		discrLevels[j] = fe.profile.DiscrLevel
		deviceIDs[j] = fe.profile.DeviceID
		nCrates[j] = int32(fe.profile.NCrate)
		dataOffsets[j] = int32(fe.dataStart)
		dataCounts[j] = int32(len(fe.profile.Data))
	}

	dimNprofiles := []string{"nprofiles"}
	profVars := []struct {
		name   string
		values any
	}{
		{"file_index", fileIdxs},
		{"active", actives},
		{"photon", photons},
		{"laser_type", laserTypes},
		{"n_data_points", nDataPoints},
		{"reserved_0", reserved0},
		{"reserved_1", reserved1},
		{"reserved_2", reserved2},
		{"high_voltage", highVoltages},
		{"bin_width", binWidths},
		{"wavelength", wavelengths},
		{"polarization", polarizations},
		{"bin_shift", binShifts},
		{"dec_bin_shift", decBinShifts},
		{"adc_bits", adcBits},
		{"n_shots", nShots},
		{"discr_level", discrLevels},
		{"device_id", deviceIDs},
		{"n_crate", nCrates},
		{"data_offset", dataOffsets},
		{"data_count", dataCounts},
	}
	for _, v := range profVars {
		if err := cw.AddVar(v.name, api.Variable{
			Values:     v.values,
			Dimensions: dimNprofiles,
		}); err != nil {
			return fmt.Errorf("adding variable %q: %w", v.name, err)
		}
	}

	// --- Data variable ---
	if err := cw.AddVar("profile_data", api.Variable{
		Values:     dataBuf,
		Dimensions: []string{"ndata"},
	}); err != nil {
		return fmt.Errorf("adding profile_data: %w", err)
	}

	return nil
}

// LoadLicelPackFromNetCDF3 загружает LicelPack из NetCDF3 (CDF) файла.
func LoadLicelPackFromNetCDF3(fname string) (*LicelPack, error) {
	nc, err := netcdf.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("opening netcdf3 file: %w", err)
	}
	defer nc.Close()

	// --- Global attributes ---
	attrs := nc.Attributes()
	startTime := parseAttrTime(attrs, "start_time")
	stopTime := parseAttrTime(attrs, "stop_time")

	// --- Dimensions ---
	nfiles, ok := nc.GetDimension("nfiles")
	if !ok {
		// Файл без размерностей — пустой пак
		return &LicelPack{
			StartTime: startTime,
			StopTime:  stopTime,
			Data:      make(map[string]LicelFile),
		}, nil
	}
	nprofiles, ok := nc.GetDimension("nprofiles")
	if !ok {
		return nil, fmt.Errorf("netcdf3 file: missing dimension \"nprofiles\"")
	}
	ndata, ok := nc.GetDimension("ndata")
	if !ok {
		return nil, fmt.Errorf("netcdf3 file: missing dimension \"ndata\"")
	}
	_ = ndata // used indirectly via profile_data

	// --- Read file-level variables ---
	filenames := readStrings(nc, "filename", int(nfiles))
	sites := readStrings(nc, "measurement_site", int(nfiles))
	startTimeStrs := readStrings(nc, "measurement_start_time", int(nfiles))
	stopTimeStrs := readStrings(nc, "measurement_stop_time", int(nfiles))
	altitudes := readFloat64s(nc, "altitude", int(nfiles))
	longitudes := readFloat64s(nc, "longitude", int(nfiles))
	latitudes := readFloat64s(nc, "latitude", int(nfiles))
	zeniths := readFloat64s(nc, "zenith", int(nfiles))
	l1ns := readInt32s(nc, "laser1_nshots", int(nfiles))
	l1f := readInt32s(nc, "laser1_freq", int(nfiles))
	l2ns := readInt32s(nc, "laser2_nshots", int(nfiles))
	l2f := readInt32s(nc, "laser2_freq", int(nfiles))
	l3ns := readInt32s(nc, "laser3_nshots", int(nfiles))
	l3f := readInt32s(nc, "laser3_freq", int(nfiles))
	ndss := readInt32s(nc, "ndatasets", int(nfiles))

	// --- Read profile-level variables ---
	fileIdxs := readInt32s(nc, "file_index", int(nprofiles))
	actives := readInt32s(nc, "active", int(nprofiles))
	photons := readInt32s(nc, "photon", int(nprofiles))
	laserTypes := readInt32s(nc, "laser_type", int(nprofiles))
	nDataPoints := readInt32s(nc, "n_data_points", int(nprofiles))
	reserved0 := readInt32s(nc, "reserved_0", int(nprofiles))
	reserved1 := readInt32s(nc, "reserved_1", int(nprofiles))
	reserved2 := readInt32s(nc, "reserved_2", int(nprofiles))
	highVoltages := readInt32s(nc, "high_voltage", int(nprofiles))
	binWidths := readFloat64s(nc, "bin_width", int(nprofiles))
	wavelengths := readFloat64s(nc, "wavelength", int(nprofiles))
	polarizations := readStrings(nc, "polarization", int(nprofiles))
	binShifts := readInt32s(nc, "bin_shift", int(nprofiles))
	decBinShifts := readInt32s(nc, "dec_bin_shift", int(nprofiles))
	adcBits := readInt32s(nc, "adc_bits", int(nprofiles))
	nShots := readInt32s(nc, "n_shots", int(nprofiles))
	discrLevels := readFloat64s(nc, "discr_level", int(nprofiles))
	deviceIDs := readStrings(nc, "device_id", int(nprofiles))
	nCrates := readInt32s(nc, "n_crate", int(nprofiles))
	dataOffsets := readInt32s(nc, "data_offset", int(nprofiles))
	dataCounts := readInt32s(nc, "data_count", int(nprofiles))

	// --- Read profile data ---
	profileData := readFloat64s(nc, "profile_data", int(ndata))

	// --- Build LicelPack ---
	// Создаём LicelFile для каждого индекса файла
	fileMap := make(map[int32]*LicelFile)
	var fileOrder []int32 // сохраняем порядок обхода

	for fi := int32(0); fi < int32(nfiles); fi++ {
		i := int(fi)
		fts := parseTimeSafe(startTimeStrs[i])
		fto := parseTimeSafe(stopTimeStrs[i])

		lf := LicelFile{
			MeasurementSite:       sites[i],
			MeasurementStartTime:  fts,
			MeasurementStopTime:   fto,
			AltitudeAboveSeaLevel: altitudes[i],
			Longitude:             longitudes[i],
			Latitude:              latitudes[i],
			Zenith:                zeniths[i],
			Laser1NShots:          int(l1ns[i]),
			Laser1Freq:            int(l1f[i]),
			Laser2NShots:          int(l2ns[i]),
			Laser2Freq:            int(l2f[i]),
			Laser3NShots:          int(l3ns[i]),
			Laser3Freq:            int(l3f[i]),
			NDatasets:             int(ndss[i]),
			FileLoaded:            true,
			Profiles:              make(LicelProfilesList, 0),
		}
		fileMap[fi] = &lf
		fileOrder = append(fileOrder, fi)
	}

	// Заполняем профили
	for j := 0; j < int(nprofiles); j++ {
		fi := fileIdxs[j]
		lf, ok := fileMap[fi]
		if !ok {
			continue
		}

		offset := int(dataOffsets[j])
		count := int(dataCounts[j])
		var data []float64
		if offset >= 0 && count > 0 && offset+count <= len(profileData) {
			data = make([]float64, count)
			copy(data, profileData[offset:offset+count])
		} else if count == 0 {
			data = []float64{}
		}

		pr := LicelProfile{
			Active:       actives[j] != 0,
			Photon:       photons[j] != 0,
			LaserType:    int(laserTypes[j]),
			NDataPoints:  int(nDataPoints[j]),
			Reserved:     [3]int{int(reserved0[j]), int(reserved1[j]), int(reserved2[j])},
			HighVoltage:  int(highVoltages[j]),
			BinWidth:     binWidths[j],
			Wavelength:   wavelengths[j],
			Polarization: polarizations[j],
			BinShift:     int(binShifts[j]),
			DecBinShift:  int(decBinShifts[j]),
			AdcBits:      int(adcBits[j]),
			NShots:       int(nShots[j]),
			DiscrLevel:   discrLevels[j],
			DeviceID:     deviceIDs[j],
			NCrate:       int(nCrates[j]),
			Data:         data,
		}
		lf.Profiles = append(lf.Profiles, pr)
	}

	// Пересчитываем NDatasets и собираем результат
	result := &LicelPack{
		StartTime: startTime,
		StopTime:  stopTime,
		Data:      make(map[string]LicelFile, len(fileMap)),
	}
	for fi, lf := range fileMap {
		lf.NDatasets = len(lf.Profiles)
		fname := filenames[int(fi)]
		result.Data[fname] = *lf
	}

	return result, nil
}

// --- helpers ---

// btoi32 converts bool to int32 (0/1).
func btoi32(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

// parseAttrTime reads a string attribute and parses it as RFC3339.
func parseAttrTime(attrs api.AttributeMap, key string) time.Time {
	v, ok := attrs.Get(key)
	if !ok {
		return time.Time{}
	}
	s, ok := v.(string)
	if !ok {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// parseTimeSafe parses an RFC3339 string, returning zero time on error.
func parseTimeSafe(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// readStrings reads a string variable from the NetCDF group.
func readStrings(g api.Group, name string, size int) []string {
	vr, err := g.GetVariable(name)
	if err != nil || vr == nil {
		return make([]string, size)
	}
	v, ok := vr.Values.([]string)
	if !ok {
		return make([]string, size)
	}
	return v
}

// readFloat64s reads a float64 variable from the NetCDF group.
func readFloat64s(g api.Group, name string, size int) []float64 {
	vr, err := g.GetVariable(name)
	if err != nil || vr == nil {
		return make([]float64, size)
	}
	v, ok := vr.Values.([]float64)
	if !ok {
		return make([]float64, size)
	}
	return v
}

// readInt32s reads an int32 variable from the NetCDF group.
func readInt32s(g api.Group, name string, size int) []int32 {
	vr, err := g.GetVariable(name)
	if err != nil || vr == nil {
		return make([]int32, size)
	}
	v, ok := vr.Values.([]int32)
	if !ok {
		return make([]int32, size)
	}
	return v
}
