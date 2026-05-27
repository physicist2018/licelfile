# v2.0.0 — refactor/licelformat

## Changelog

## [v2.1.1] — 2026-05-25

### Changed

- **`SelectProfile(isPhoton, wavelength, polarization)`** — добавлен параметр `polarization string`. Передайте `""` чтобы подходила любая поляризация.
- **`SelectProfiles(isPhoton, wavelength, polarization)`** — аналогично, добавлен параметр `polarization string`.

### Added

- Тесты поляризации: `TestLicelFile_SelectProfile` (match by pol, mismatch, any), `TestSelectProfiles` (2 files, multiple polarizations).

---

## [v2.1.0] — 2026-05-25

### Added

- **`LicelPack.ZipCompressionLevel`** — поле для управления степенью сжатия zip-архива (0 = deflate по умолчанию, 1–9 = степень deflate).
- **`writeCompressedEntry`** — вспомогательная функция для создания zip-записей с заданным уровнем сжатия через `CreateRaw`.
- Тест **`TestLicelPack_SaveToZip_CompressionLevels`** — проверяет round-trip для уровней 0, 1, 5, 9.

### Changed

- **`SaveToZip`**: использует `writeCompressedEntry` при `ZipCompressionLevel > 0`, иначе стандартный `zw.Create`.

---

## [v2.0.2] — 2026-05-25

### Changed

- **`parseTime`**: `time.UTC` → `time.Local` — времена теперь парсятся и форматируются в локальном часовом поясе.
- **`FormatSecondLine`**: `.UTC().Format()` → `.Local().Format()` — консистентно с `parseTime`.
- **`isValidFilename`**: regex `^b.*\..+` → `^[a-z].*\..+` — файлы Licel теперь могут начинаться на любую букву `[a-z]`.

### Fixed

- **Тест `TestParseTime`**: убран вызов `.UTC()`, проверяется локальное время напрямую.
- **Тест `TestIsValidFilename`**: `a12345.678901` теперь валидный.

---

## [v2.0.1] — 2026-05-25

### Removed

- **`timefmt` dependency** — парсинг дат переведён на стандартный `time.ParseInLocation`.
- **`zerolog` dependency** — полностью удалена из `go.mod` (не использовалась после v2.0.0).

### Changed

- **`parseTime`**: `timefmt.Strptime` → `time.ParseInLocation("02/01/2006 15:04:05", s, time.UTC)`.
- **`FormatSecondLine`**: `.Format()` → `.UTC().Format()` для консистентности.

### Added

- Тест `TestParseTime_WrongOrder` — проверка, что MM/DD/YYYY не сработает вместо DD/MM/YYYY.

---

## [v2.0.0] — 2026-05-25

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
