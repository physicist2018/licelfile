# Changelog

## [Unreleased] — refactor/licelformat

### Breaking changes

- **`LoadLicelFile`** — returns `(LicelFile, error)` instead of `LicelFile`.
- **`LoadLicelFileFromReader`** — returns `(LicelFile, error)` instead of `LicelFile`; unused `size` parameter removed.
- **`NewLicelPack`** — returns `(*LicelPack, error)` instead of `*LicelPack`.
- **`NewLicelPackFromZip`** — returns `(*LicelPack, error)` instead of `*LicelPack`.
- **`NewLicelProfile`** — returns `(LicelProfile, error)` instead of `LicelProfile`.
- **`SelectCertainWavelength1`** — renamed to **`SelectProfile`**, returns `(LicelProfile, bool)` instead of `LicelProfile`.
- **`SelectCertainWavelength2`** — renamed to **`SelectProfiles`**, operates on `*LicelPack` value receiver.
- **`LicelProfile.Profile()`** — returns `([]byte, error)` instead of `string`.
- **`LicelProfilesList`** — moved from `licelfile.go` to `licelprofile.go` (same package, no import change).

### Removed

- **zerolog** dependency removed from the library. Errors are returned to the caller instead of calling `log.Fatal()`.
- `SelectCertainWavelength1`, `SelectCertainWavelength2` — replaced by `SelectProfile` / `SelectProfiles`.

### Fixed

- **round-trip save/load**: `Save()` previously wrote already-scaled data, causing double-scaling on reload. Fixed via `ProfileRaw()` method.
- **`Save()` filename bug**: no longer appends `"1"` to the filename.
- **`Metadata()` reserved fields**: `Reserved[1]` and `Reserved[2]` are now serialized instead of hardcoded `0, 0`.
- **`str2Bool`, `str2Int`, `str2Float`**: no longer silently swallow parse errors; all return `error`.
- **`readAndTrimLine`, `skipCRLF`, `parseTime`**: no longer call `log.Fatal()`; errors are returned.

### Added

- **`LicelFile.WriteTo(w io.Writer, fname string) error`** — serialize directly into any `io.Writer`.
- **`LicelProfile.scaleFactor() float64`** — computes the channel's scale factor.
- **`LicelProfile.ProfileRaw() ([]byte, error)`** — returns unscaled binary data (for safe save/reload round-trips).
- **`LicelPack.SaveToZip(zipPath string) error`** — saves all pack files into a zip archive.
- **30 unit tests** across `licelfile_test.go`, `licelprofile_test.go`, `licelpack_test.go`.
- **Validation**: `NewLicelProfile` checks field count, `wavelength.polarization` format, and wraps all parse errors.

### Changed

- **`Save()`** uses `bufio.Writer` and delegates to `WriteTo()`.
- **`loadFromReader`** eliminates ~80% code duplication between `LoadLicelFile` and `LoadLicelFileFromReader`.
- **`isValidFilename`** uses a pre-compiled `regexp` instead of compiling on each call.
