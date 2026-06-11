package licelformat

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/batchatco/go-native-netcdf/netcdf"
	"github.com/batchatco/go-native-netcdf/netcdf/api"
	"github.com/batchatco/go-native-netcdf/netcdf/util"
)

// ─── helpers for building variable attributes ───────────────────────────────

type attrBuilder struct {
	keys []string
	vals map[string]any
}

func newAttrs() *attrBuilder {
	return &attrBuilder{vals: make(map[string]any)}
}

func (ab *attrBuilder) add(k string, v any) *attrBuilder {
	ab.keys = append(ab.keys, k)
	ab.vals[k] = v
	return ab
}

func (ab *attrBuilder) build() (api.AttributeMap, error) {
	if len(ab.keys) == 0 {
		return nil, nil
	}
	return util.NewOrderedMap(ab.keys, ab.vals)
}

// ─── SaveToNetCDF3 – saves LicelPack as NetCDF3 (CDF 64-bit) ────────────────
//
// Schema (CF-1.8, matching Python reference):
//
//	Global attrs: Conventions, source
//	Dimensions:   file, profile, range
//	Coordinate:   range (float64, meters)
//	File vars:    file_name, site, start_time, stop_time, longitude, latitude,
//	              altitude, zenith, laser{1,2,3}_{nshots,freq}, ndatasets
//	Profile vars: file_index, wavelength, polarization, bin_width, nshots,
//	              device_id, is_photon, discr_level, adc_bits, active,
//	              laser_type, high_voltage, bin_shift, dec_bin_shift,
//	              n_crate, npoints
//	Data:         signal (profile × range, float64, NaN-padded)
func (lp *LicelPack) SaveToNetCDF3(fname string) error {
	nfiles := len(lp.Data)
	if nfiles == 0 {
		return fmt.Errorf("cannot save an empty LicelPack to NetCDF")
	}

	// Sorted keys for deterministic ordering.
	fileKeys := make([]string, 0, nfiles)
	for k := range lp.Data {
		fileKeys = append(fileKeys, k)
	}
	sort.Strings(fileKeys)

	// Build flat profile list and determine max range.
	type flatEntry struct {
		fileIdx int
		profile LicelProfile
	}
	var flat []flatEntry
	maxRange := 0

	for fIdx, k := range fileKeys {
		lf := lp.Data[k]
		for _, pr := range lf.Profiles {
			flat = append(flat, flatEntry{fileIdx: fIdx, profile: pr})
			if pr.NDataPoints > maxRange {
				maxRange = pr.NDataPoints
			}
		}
	}

	nprofiles := len(flat)
	if nprofiles == 0 {
		return fmt.Errorf("no profiles in the pack")
	}

	cw, err := netcdf.OpenWriter(fname, netcdf.KindCDF)
	if err != nil {
		return fmt.Errorf("creating netcdf3 file: %w", err)
	}
	defer cw.Close()

	// --- Global attributes ---
	ga, err := newAttrs().
		add("Conventions", "CF-1.8").
		add("source", "licelformat v1").
		add("licelformat_version", int32(1)).
		build()
	if err != nil {
		return fmt.Errorf("global attributes: %w", err)
	}
	if ga != nil {
		if err := cw.AddAttributes(ga); err != nil {
			return fmt.Errorf("add global attributes: %w", err)
		}
	}

	// --- File-level slices ---
	filenames := make([]string, nfiles)
	sites := make([]string, nfiles)
	startTimes := make([]float64, nfiles)
	stopTimes := make([]float64, nfiles)
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

	for i, k := range fileKeys {
		lf := lp.Data[k]
		filenames[i] = k
		sites[i] = lf.MeasurementSite
		startTimes[i] = float64(lf.MeasurementStartTime.Unix())
		stopTimes[i] = float64(lf.MeasurementStopTime.Unix())
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
	}

	// --- Profile-level slices ---
	fileIdxs := make([]int32, nprofiles)
	wavelengths := make([]float64, nprofiles)
	polarizations := make([]string, nprofiles)
	binWidths := make([]float64, nprofiles)
	nshots := make([]int32, nprofiles)
	deviceIDs := make([]string, nprofiles)
	isPhotons := make([]int32, nprofiles)
	discrLevels := make([]float64, nprofiles)
	adcBits := make([]int32, nprofiles)
	actives := make([]int32, nprofiles)
	laserTypes := make([]int32, nprofiles)
	highVoltages := make([]int32, nprofiles)
	binShifts := make([]int32, nprofiles)
	decBinShifts := make([]int32, nprofiles)
	nCrates := make([]int32, nprofiles)
	npoints := make([]int32, nprofiles)
	reserved0 := make([]int32, nprofiles)
	reserved1 := make([]int32, nprofiles)
	reserved2 := make([]int32, nprofiles)

	for j, fe := range flat {
		fileIdxs[j] = int32(fe.fileIdx)
		wavelengths[j] = fe.profile.Wavelength
		polarizations[j] = fe.profile.Polarization
		binWidths[j] = fe.profile.BinWidth
		nshots[j] = int32(fe.profile.NShots)
		deviceIDs[j] = fe.profile.DeviceID
		isPhotons[j] = btoi32(fe.profile.Photon)
		discrLevels[j] = fe.profile.DiscrLevel
		adcBits[j] = int32(fe.profile.AdcBits)
		actives[j] = btoi32(fe.profile.Active)
		laserTypes[j] = int32(fe.profile.LaserType)
		highVoltages[j] = int32(fe.profile.HighVoltage)
		binShifts[j] = int32(fe.profile.BinShift)
		decBinShifts[j] = int32(fe.profile.DecBinShift)
		nCrates[j] = int32(fe.profile.NCrate)
		npoints[j] = int32(fe.profile.NDataPoints)
		reserved0[j] = int32(fe.profile.Reserved[0])
		reserved1[j] = int32(fe.profile.Reserved[1])
		reserved2[j] = int32(fe.profile.Reserved[2])
	}

	// --- Signal: 2D NaN-padded ---
	signal := make([][]float64, nprofiles)
	for j, fe := range flat {
		row := make([]float64, maxRange)
		for k := range row {
			row[k] = math.NaN()
		}
		for k, v := range fe.profile.Data {
			row[k] = v
		}
		signal[j] = row
	}

	// --- Range coordinate ---
	firstBW := float64(7.5)
	if len(flat) > 0 {
		firstBW = flat[0].profile.BinWidth
	}
	rangeVals := make([]float64, maxRange)
	for k := 0; k < maxRange; k++ {
		rangeVals[k] = float64(k) * firstBW
	}

	dimFile := []string{"file"}
	dimProfile := []string{"profile"}
	dimRange := []string{"range"}
	dimSignal := []string{"profile", "range"}

	// ── File-level variables ──────────────────────────────────────────────────

	if err := addStrVar(cw, "file_name", filenames, dimFile, "original file name"); err != nil {
		return err
	}
	if err := addStrVar(cw, "site", sites, dimFile, "measurement site"); err != nil {
		return err
	}
	if err := addFloatVarWithCalendar(cw, "start_time", startTimes, dimFile,
		"measurement start time", "seconds since 1970-01-01 00:00:00 UTC", "standard", math.NaN()); err != nil {
		return err
	}
	if err := addFloatVarWithCalendar(cw, "stop_time", stopTimes, dimFile,
		"measurement stop time", "seconds since 1970-01-01 00:00:00 UTC", "standard", math.NaN()); err != nil {
		return err
	}
	if err := addFloatVar(cw, "longitude", longitudes, dimFile, "longitude", "degrees_east", math.NaN()); err != nil {
		return err
	}
	if err := addFloatVar(cw, "latitude", latitudes, dimFile, "latitude", "degrees_north", math.NaN()); err != nil {
		return err
	}
	if err := addFloatVar(cw, "altitude", altitudes, dimFile, "lidar altitude above sea level", "meters", math.NaN()); err != nil {
		return err
	}
	if err := addFloatVar(cw, "zenith", zeniths, dimFile, "zenith angle", "degrees", math.NaN()); err != nil {
		return err
	}
	if err := addIntVar(cw, "laser1_nshots", l1ns, dimFile, "laser 1 number of shots", "", int32(-1)); err != nil {
		return err
	}
	if err := addIntVar(cw, "laser1_freq", l1f, dimFile, "laser 1 frequency", "Hz", int32(-1)); err != nil {
		return err
	}
	if err := addIntVar(cw, "laser2_nshots", l2ns, dimFile, "laser 2 number of shots", "", int32(-1)); err != nil {
		return err
	}
	if err := addIntVar(cw, "laser2_freq", l2f, dimFile, "laser 2 frequency", "Hz", int32(-1)); err != nil {
		return err
	}
	if err := addIntVar(cw, "laser3_nshots", l3ns, dimFile, "laser 3 number of shots", "", int32(-1)); err != nil {
		return err
	}
	if err := addIntVar(cw, "laser3_freq", l3f, dimFile, "laser 3 frequency", "Hz", int32(-1)); err != nil {
		return err
	}
	if err := addIntVar(cw, "ndatasets", ndss, dimFile, "number of datasets (profiles) per file", "", int32(-1)); err != nil {
		return err
	}

	// ── Profile-level variables ───────────────────────────────────────────────

	if err := addIntVar(cw, "file_index", fileIdxs, dimProfile, "index of the parent file", "", int32(-1)); err != nil {
		return err
	}
	if err := addFloatVar(cw, "wavelength", wavelengths, dimProfile, "laser wavelength", "nanometers", math.NaN()); err != nil {
		return err
	}
	if err := addStrVar(cw, "polarization", polarizations, dimProfile, "polarization channel"); err != nil {
		return err
	}
	if err := addFloatVar(cw, "bin_width", binWidths, dimProfile, "range bin width", "meters", math.NaN()); err != nil {
		return err
	}
	if err := addIntVar(cw, "nshots", nshots, dimProfile, "number of laser shots", "", int32(-1)); err != nil {
		return err
	}
	if err := addStrVar(cw, "device_id", deviceIDs, dimProfile, "device identifier"); err != nil {
		return err
	}
	if err := addIntVarWithFlags(cw, "is_photon", isPhotons, dimProfile,
		"photon counting channel flag", "", int32(-1)); err != nil {
		return err
	}
	if err := addFloatVar(cw, "discr_level", discrLevels, dimProfile, "discriminator level", "millivolts", math.NaN()); err != nil {
		return err
	}
	if err := addIntVar(cw, "adc_bits", adcBits, dimProfile, "ADC resolution", "", int32(-1)); err != nil {
		return err
	}
	if err := addIntVar(cw, "active", actives, dimProfile, "channel active flag", "", int32(-1)); err != nil {
		return err
	}
	if err := addIntVar(cw, "laser_type", laserTypes, dimProfile, "laser index", "", int32(-1)); err != nil {
		return err
	}
	if err := addIntVar(cw, "high_voltage", highVoltages, dimProfile, "PMT high voltage", "volts", int32(-1)); err != nil {
		return err
	}
	if err := addIntVar(cw, "bin_shift", binShifts, dimProfile, "", "", int32(-1)); err != nil {
		return err
	}
	if err := addIntVar(cw, "dec_bin_shift", decBinShifts, dimProfile, "", "", int32(-1)); err != nil {
		return err
	}
	if err := addIntVar(cw, "n_crate", nCrates, dimProfile, "", "", int32(-1)); err != nil {
		return err
	}
	if err := addIntVar(cw, "reserved_0", reserved0, dimProfile, "", "", int32(-1)); err != nil {
		return err
	}
	if err := addIntVar(cw, "reserved_1", reserved1, dimProfile, "", "", int32(-1)); err != nil {
		return err
	}
	if err := addIntVar(cw, "reserved_2", reserved2, dimProfile, "", "", int32(-1)); err != nil {
		return err
	}
	if err := addIntVar(cw, "npoints", npoints, dimProfile, "number of valid data points", "", int32(-1)); err != nil {
		return err
	}

	// ── Range coordinate ──────────────────────────────────────────────────────

	if err := addFloatVar(cw, "range", rangeVals, dimRange, "range from lidar", "meters", math.NaN()); err != nil {
		return err
	}

	// ── Signal (2D) ───────────────────────────────────────────────────────────

	attrs, err := newAttrs().add("long_name", "lidar signal").add("units", "millivolts").add("FillValue", math.NaN()).add("cell_methods", "range: mean").build()
	if err != nil {
		return fmt.Errorf("signal attrs: %w", err)
	}
	if err := cw.AddVar("signal", api.Variable{
		Values:     signal,
		Dimensions: dimSignal,
		Attributes: attrs,
	}); err != nil {
		return err
	}

	return nil
}

// ─── variable-adding helpers ─────────────────────────────────────────────────

func addStrVar(cw api.Writer, name string, values []string, dims []string, longName string) error {
	attrs, err := newAttrs().add("long_name", longName).build()
	if err != nil {
		return fmt.Errorf("%s attrs: %w", name, err)
	}
	return cw.AddVar(name, api.Variable{
		Values:     values,
		Dimensions: dims,
		Attributes: attrs,
	})
}

func addFloatVar(cw api.Writer, name string, values []float64, dims []string, longName, units string, fillValue float64) error {
	ab := newAttrs().add("long_name", longName)
	if units != "" {
		ab = ab.add("units", units)
	}
	ab = ab.add("FillValue", fillValue)
	attrs, err := ab.build()
	if err != nil {
		return fmt.Errorf("%s attrs: %w", name, err)
	}
	return cw.AddVar(name, api.Variable{
		Values:     values,
		Dimensions: dims,
		Attributes: attrs,
	})
}

func addIntVar(cw api.Writer, name string, values []int32, dims []string, longName, units string, fillValue int32) error {
	ab := newAttrs().add("long_name", longName)
	if units != "" {
		ab = ab.add("units", units)
	}
	ab = ab.add("FillValue", fillValue)
	attrs, err := ab.build()
	if err != nil {
		return fmt.Errorf("%s attrs: %w", name, err)
	}
	return cw.AddVar(name, api.Variable{
		Values:     values,
		Dimensions: dims,
		Attributes: attrs,
	})
}

func addFloatVarWithCalendar(cw api.Writer, name string, values []float64, dims []string, longName, units, calendar string, fillValue float64) error {
	ab := newAttrs().add("long_name", longName).add("units", units).add("calendar", calendar).add("FillValue", fillValue)
	attrs, err := ab.build()
	if err != nil {
		return fmt.Errorf("%s attrs: %w", name, err)
	}
	return cw.AddVar(name, api.Variable{
		Values:     values,
		Dimensions: dims,
		Attributes: attrs,
	})
}

func addIntVarWithFlags(cw api.Writer, name string, values []int32, dims []string, longName, units string, fillValue int32) error {
	ab := newAttrs().add("long_name", longName)
	if units != "" {
		ab = ab.add("units", units)
	}
	ab = ab.add("FillValue", fillValue)
	ab = ab.add("flag_values", "0, 1")
	ab = ab.add("flag_meanings", "analog photon_counting")
	attrs, err := ab.build()
	if err != nil {
		return fmt.Errorf("%s attrs: %w", name, err)
	}
	return cw.AddVar(name, api.Variable{
		Values:     values,
		Dimensions: dims,
		Attributes: attrs,
	})
}

// ─── LoadLicelPackFromNetCDF3 – loads from NetCDF3 (CDF) ─────────────────────

func LoadLicelPackFromNetCDF3(fname string) (*LicelPack, error) {
	nc, err := netcdf.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("opening netcdf3 file: %w", err)
	}
	defer nc.Close()

	// --- Read dimensions ---
	nfiles, ok := nc.GetDimension("file")
	if !ok || nfiles == 0 {
		return &LicelPack{
			Data: make(map[string]LicelFile),
		}, nil
	}
	nprofiles, ok := nc.GetDimension("profile")
	if !ok {
		return nil, fmt.Errorf("missing dimension \"profile\"")
	}

	// --- Read file-level variables ---
	filenames := readStrings(nc, "file_name", int(nfiles))
	sites := readStrings(nc, "site", int(nfiles))
	startTimes := readFloat64s(nc, "start_time", int(nfiles))
	stopTimes := readFloat64s(nc, "stop_time", int(nfiles))
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
	wavelengths := readFloat64s(nc, "wavelength", int(nprofiles))
	polarizations := readStrings(nc, "polarization", int(nprofiles))
	binWidths := readFloat64s(nc, "bin_width", int(nprofiles))
	nshots := readInt32s(nc, "nshots", int(nprofiles))
	deviceIDs := readStrings(nc, "device_id", int(nprofiles))
	isPhotons := readInt32s(nc, "is_photon", int(nprofiles))
	discrLevels := readFloat64s(nc, "discr_level", int(nprofiles))
	adcBits := readInt32s(nc, "adc_bits", int(nprofiles))
	actives := readInt32s(nc, "active", int(nprofiles))
	laserTypes := readInt32s(nc, "laser_type", int(nprofiles))
	highVoltages := readInt32s(nc, "high_voltage", int(nprofiles))
	binShifts := readInt32s(nc, "bin_shift", int(nprofiles))
	decBinShifts := readInt32s(nc, "dec_bin_shift", int(nprofiles))
	nCrates := readInt32s(nc, "n_crate", int(nprofiles))
	npoints := readInt32s(nc, "npoints", int(nprofiles))
	reserved0 := readInt32s(nc, "reserved_0", int(nprofiles))
	reserved1 := readInt32s(nc, "reserved_1", int(nprofiles))
	reserved2 := readInt32s(nc, "reserved_2", int(nprofiles))

	// --- Read signal (2D: profile × range) ---
	signalVar, err := nc.GetVariable("signal")
	if err != nil {
		return nil, fmt.Errorf("reading signal: %w", err)
	}
	signalRows, ok := signalVar.Values.([][]float64)
	if !ok {
		return nil, fmt.Errorf("signal variable is not [][]float64, got %T", signalVar.Values)
	}

	// --- Build LicelPack ---
	fileMap := make(map[int32]*LicelFile)
	var fileOrder []int32

	for fi := int32(0); fi < int32(nfiles); fi++ {
		i := int(fi)
		fts := time.Unix(int64(startTimes[i]), 0)
		fto := time.Unix(int64(stopTimes[i]), 0)

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

	// Fill profiles from 2D signal
	for j := 0; j < int(nprofiles); j++ {
		fi := fileIdxs[j]
		lf, ok := fileMap[fi]
		if !ok {
			continue
		}

		np := int(npoints[j])
		data := make([]float64, np)
		if j < len(signalRows) {
			for k := 0; k < np && k < len(signalRows[j]); k++ {
				data[k] = signalRows[j][k]
			}
		}

		pr := LicelProfile{
			Active:       actives[j] != 0,
			Photon:       isPhotons[j] != 0,
			LaserType:    int(laserTypes[j]),
			NDataPoints:  np,
			Reserved:     [3]int{int(reserved0[j]), int(reserved1[j]), int(reserved2[j])},
			HighVoltage:  int(highVoltages[j]),
			BinWidth:     binWidths[j],
			Wavelength:   wavelengths[j],
			Polarization: polarizations[j],
			BinShift:     int(binShifts[j]),
			DecBinShift:  int(decBinShifts[j]),
			AdcBits:      int(adcBits[j]),
			NShots:       int(nshots[j]),
			DiscrLevel:   discrLevels[j],
			DeviceID:     deviceIDs[j],
			NCrate:       int(nCrates[j]),
			Data:         data,
		}
		lf.Profiles = append(lf.Profiles, pr)
	}

	// Compute pack StartTime / StopTime from file min/max
	var packStart, packStop time.Time
	for _, lf := range fileMap {
		if packStart.IsZero() || lf.MeasurementStartTime.Before(packStart) {
			packStart = lf.MeasurementStartTime
		}
		if packStop.IsZero() || lf.MeasurementStopTime.After(packStop) {
			packStop = lf.MeasurementStopTime
		}
	}

	result := &LicelPack{
		StartTime: packStart,
		StopTime:  packStop,
		Data:      make(map[string]LicelFile, len(fileMap)),
	}
	for fi, lf := range fileMap {
		lf.NDatasets = len(lf.Profiles)
		fname := filenames[int(fi)]
		result.Data[fname] = *lf
	}

	return result, nil
}

// ─── utility helpers ─────────────────────────────────────────────────────────

func btoi32(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

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
