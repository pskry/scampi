# Changelog

## v0.1.0-alpha.2 — 2026-03-18

### Features
- Auto-escalate to sudo on permission errors (#46)
- Source resolvers: unified file acquisition for all steps (#55)
- Source resolver: remote() (#56)

### Enhancements
- scampi index: wrap step descriptions instead of truncating (#36)
- Generalize promised resources beyond paths (#43)

### Bug Fixes
- Graceful cancellation on SIGINT instead of panic (#45)

## v0.1.0-alpha.1 — 2026-03-17

### Features
- Step: sysctl (#19)
- Step: firewall (#20)
- Verify field for copy and template steps (#35)
- Step: unarchive (#40)
- Automated site deployment on release (#48)
- User/group step (#5)
- Post-change hooks (#7)

### Enhancements
- Action-started feedback (#10)
- Service reload/restart (#11)
- Error message consistency pass (#16)
- Unify three copy-pasted cycle detection implementations (#32)
- scampi index should show default values for optional fields (#33)
- Inline content for copy step (symmetry with template) (#34)
- Add benchmark and fuzz coverage for all step types (#37)
- Proper action dependency system (#38)
- Deduplicate engine test fixtures (#39)

### Bug Fixes
- scampi inspect doesn't show template steps (#14)
- Fix goroutine leak in benchmark suite (#42)
- Check across uncommitted changes (#9)

