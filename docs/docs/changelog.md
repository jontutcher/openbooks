## [Development]

### Added

- The IRC connection is kept open between page reloads / client disconnects.
    - After the client disconnects, the IRC connection will remain open until the client reconnects or 3 minutes have passed.
      This will prevent inadvertent bans from the IRC channel due to frequent connects/disconnects.
- Grid Improvements
    - Column sorting
    - Fuzzy searching across title / author columns
    - Toggle column visibility
- A `Test` GitHub Actions workflow that lints the client and runs `gofmt`, `go vet` and `go test`.

### Changed

- Updated the Go toolchain to 1.26 and every Go dependency to its current release.
- Replaced the archived `mholt/archiver/v3` with `mholt/archives`, and `webview/webview` with `webview/webview_go`.
- Updated the client to React 19, Mantine 9, TanStack Table 9, Redux Toolkit 2, Vite 8 and ESLint 9 (flat config).
- Updated the Docker build, the GitHub Actions workflows, the dev container and the documentation toolchain.
- `cmd/mock` now takes `-server`, `-search-results` and `-book-result` flags instead of hard coding them, and `task dev:irc-stop` removes the development IRC server.
- The build script has been removed in favour of `task release:build`, which the release workflow already used.

### Fixed

- Search results with no author in the file name are no longer discarded. Names
  such as `!peapod The Great Gatsby.pdf` or `!DeathCookie Travis_Bagwell_Tarot_03_Inferno.epub.rar`
  are now listed with an empty author instead of counting as parse errors.
- File sizes that a server separates from their unit (`205.10 KiB`) keep the unit.
- Table text and the facet filter counts are no longer near-invisible in dark mode.
- The library badge renders again; it referred to a `brand` colour the theme does not define.
- Restored the column facet filter, the empty-state heading and the close button inside notifications, whose styles were missing.

## [v4.5.0] - 2023-01-08

### Added

- Use `--useragent/-u` flag to optionally specify the [UserAgent](https://en.wikipedia.org/wiki/Client-to-client_protocol#VERSION) reported to the IRC server. Default remains `OpenBooks v4.5.0`.

### Breaking

- `--name/-n` flag **must** be specified when starting the application. OpenBooks will no longer generate a random `noun_adjective` username.
- Only a single connection to the IRC server will be made. Opening a second browser tab will show an error message.
