# v2.0.0 — refactor/licelformat

## Changelog

## [v2.4.0] — 2026-06-04

### Added

- **`LicelProfile.IsPhoton() bool`** — возвращает `true`, если `DeviceID == "BC"` (фотонный канал).
- **`LicelProfile.IsAnalog() bool`** — возвращает `true`, если `DeviceID == "BT"` (аналоговый канал).
- **`LicelProfile.IsGlued() bool`** — возвращает `true`, если `DeviceID == "BG"` (склеенный канал).
- **Тесты**: 6 тестов `TestLicelProfile_IsPhoton_*`, `TestLicelProfile_IsAnalog_*`, `TestLicelProfile_IsGlued_*`.

---

## [v2.3.0] — 2026-06-04

### Added

- **`LicelFile.Glue(wvl float64, polarization string, h1, h2 float64) (LicelProfile, error)`** — склейка аналогового и цифрового каналов для заданной длины волны и поляризации. В диапазоне высот [h1; h2] вычисляется среднее отношение analog/photon (коэффициент k). Результирующий профиль имеет DeviceID="BG": для h < h1 — данные аналогового канала; для h1 ≤ h ≤ h2 — 0.5*(analog + k*photon); для h > h2 — k*photon.
- **`LicelPack.Glue(wvl float64, polarization string, h1, h2 float64) error`** — вызывает `LicelFile.Glue` для каждого файла в паке и добавляет склеенный профиль в его `Profiles`. При ошибке в любом файле возвращает её с указанием имени файла.
- **Тесты**: 11 тестов (7 для `LicelFile.Glue_*` + 4 для `LicelPack.Glue_*`).

---

## [v2.2.0] — 2026-06-04

### Added

- **`LicelPack.ToProfilesList() LicelProfilesList`** — возвращает все профили из всех файлов пакета в виде одного плоского списка. Исходный пак не изменяется.
- **Тесты**: `TestLicelPack_ToProfilesList_*` (3 шт.).

---

## [v2.1.7] — 2026-06-03

### Added

- **`LicelPack.Filter(cond func(lf *LicelFile) bool) LicelPack`** — возвращает новый пак, содержащий только файлы, удовлетворяющие условию `cond`. Исходный пак не изменяется, `StartTime`/`StopTime` пересчитываются.
- **`LicelPack.FilterProfiles(cond func(pr *LicelProfile) bool) LicelPack`** — возвращает новый пак с отфильтрованными профилями внутри каждого файла. Файлы без подходящих профилей исключаются, `NDatasets` обновляется. Унифицирует `SelectProfile`/`SelectProfiles` под единый интерфейс.
- **`LicelPack.FilterProfilesList(cond func(pr *LicelProfile) bool) LicelProfilesList`** — возвращает объединённый список профилей из всех файлов пакета, удовлетворяющих условию `cond`. Исходный пак не изменяется.
- **Тесты**: `TestLicelPack_Filter_*` (4 шт.), `TestLicelPack_FilterProfiles_*` (4 шт.), `TestLicelPack_FilterProfilesList_*` (4 шт.).

---

## [v2.1.4] — 2026-05-30

### Added

- **`LicelProfile.SetMaxDist(alt float64) error`** — обрезает `Data` и обновляет `NDataPoints` по заданной дальности (`idx = alt / BinWidth`). Ошибка если `idx ≤ 0` или `idx > NDataPoints`.
- **`LicelFile.SetMaxDist(alt float64) error`** — вызывает `SetMaxDist` для каждого профиля в файле.
- **`LicelPack.SetMaxDist(alt float64) error`** — вызывает `SetMaxDist` для каждого файла в паке.
- **Тесты**: `TestLicelProfile_SetMaxDist`, `TestLicelProfile_SetMaxDist_ZeroAlt`, `TestLicelProfile_SetMaxDist_TooLarge`, `TestLicelProfile_SetMaxDist_ZeroBinWidth`, `TestLicelFile_SetMaxDist`, `TestLicelFile_SetMaxDist_Error`, `TestLicelPack_SetMaxDist`, `TestLicelPack_SetMaxDist_Error`.

---

## [v2.1.3] — 2026-05-30

### Added

- **`LicelPack.StopTime`** — поле `time.Time` с bson-тегом `stop_time`. Хранит время окончания последнего измерения в паке.

### Fixed

- **`NewLicelPack`** и **`NewLicelPackFromZip`**: `StartTime` теперь корректно вычисляется как минимальное время начала среди всех загруженных файлов (а не время первого). `StopTime` — максимальное время окончания.

### Changed

- **`go.mod`**: module path изменён на `github.com/physicist2018/licelfile/v2` (соответствует major version v2).
- **`AGENTS.md`**: архитектурный документ переработан — заменена гексагональная архитектура на Clean Architecture с Gin; добавлены принципы инверсии зависимостей, изоляции бизнес-логики, тестируемости; детализирован workflow добавления фич.

---

## [v2.1.2] — 2026-05-30

### Removed

- **`LicelProfile.Profile()`** — мёртвый метод, дублировал логику `float64toInt32Bytes` без учёта масштабирования. Используйте `ProfileRaw()` для получения бинарных данных с правильным unscaling.
- **Тесты `TestLicelProfile_Profile` и `TestLicelProfile_Profile_Empty`** — удалены вместе с методом.

---

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
