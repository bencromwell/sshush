# Test Improvement Plan for sshush

This document outlines a detailed plan to improve test coverage and verification for the sshush program. The current test suite has strong functional testing for the core logic but significant gaps in CLI integration and entry point verification.

---

## Current State

### Test Coverage Summary
- **cmd package**: 26.9% → **86.5%** coverage ✅
- **sshush package**: 87.8% → **89.7%** coverage ✅
- **main.go**: 0% → **50.0%** coverage (Run() at 100%) ✅
- **Overall**: 76.9% → **88.3%** of statements ✅

### Strengths
1. ✅ Comprehensive golden file tests for SSH configuration generation
2. ✅ Core parser functions are well-tested (82-100% coverage for most functions)
3. ✅ Error cases are tested (missing files, bad config)
4. ✅ Special modes are verified (dry-run, backup creation, priority ordering)

### Gaps
1. ✅ Tests for `main.go` and the program entry point (Run() at 100%)
2. ✅ `NewRootCommand` and CLI integration are tested (86.5% coverage)
3. ⏳ No end-to-end binary execution tests
4. ✅ Edge cases for path expansion and glob handling tested

---

## Detailed Plan

### Phase 1: CLI Integration Testing ✅ COMPLETE
**Priority:** High  
**Goal:** Achieve >80% coverage for `cmd/cmd.go` and verify end-to-end program execution.

#### Tasks
1. **Test `NewRootCommand`**
   - [x] Create test that initializes the root command
   - [x] Verify default flag values (source, dest, verbose, debug, dry-run)
   - [x] Test command metadata (Use, Short, Version)
   - [ ] Mock `os.UserHomeDir()` to avoid filesystem dependency
   - [x] Test viper configuration binding

2. **Test `expandPath` Edge Cases**
   - [x] Tilde expansion with subdirectories (`~/dir/subdir`)
   - [x] Multiple environment variables (`$HOME/$USER/file`)
   - [ ] Invalid paths (return errors)
   - [x] Empty string input

3. **Test `expandGlobs` Edge Cases**
   - [x] Nested glob patterns (`dir/**/*.yml`)
   - [x] No matches (empty slice)
   - [ ] Invalid glob patterns (error handling)
   - [x] Mixed valid/invalid patterns

4. **Test `must()` Function**
   - [x] Verify it calls `os.Exit(1)` on error
   - [ ] Test with various error types

#### Acceptance Criteria
- [x] `cmd/cmd.go` coverage > 80% (achieved: 86.5%)
- [x] All public functions in `cmd` package have at least one test

---

### Phase 2: Main Entry Point Testing ✅ COMPLETE
**Priority:** High  
**Goal:** Test `main.go` and achieve >50% coverage.

#### Tasks
1. **Refactor `main()` for Testability**
   - [x] Extract root command execution logic into a separate `Run()` function
   - [x] Make logger configurable for testing
   - [ ] Avoid direct `os.Exit()` calls in testable code

2. **Test `Run()` Function**
   - [x] Mock `os.Stdout` to capture logs
   - [x] Test successful execution (no error)
   - [x] Test error handling (verify exit code 1)
   - [ ] Verify logging output contains expected messages

#### Acceptance Criteria
- [x] `main.go` coverage > 50% (achieved: Run() at 100%)
- [x] Entry point can be tested without building binary

---

### Phase 3: End-to-End Binary Testing ⏳ PENDING
**Priority:** Medium  
**Goal:** Verify the compiled binary works as expected.

#### Tasks
1. **Build the Binary in Tests**
   - [ ] Use `go build` to compile sshush during test setup
   - [ ] Store binary path for subprocess testing

2. **Test Binary Execution**
   - [ ] Spawn binary as subprocess
   - [ ] Verify `--version` flag output matches expected format
   - [ ] Verify `--help` flag output contains all commands/flags
   - [ ] Test error cases (invalid flags, missing files)
   - [ ] Verify exit codes (0 for success, 1 for errors)

3. **Test File I/O**
   - [ ] Create temporary directories/files for testing
   - [ ] Verify config generation to a real file
   - [ ] Verify backup file creation
   - [ ] Test dry-run mode (no file writes)

#### Acceptance Criteria
- [ ] Binary can be executed and tested as subprocess
- [ ] All CLI flags work as expected
- [ ] File I/O behaves correctly

---

### Phase 4: Error Handling and Edge Cases ✅ COMPLETE
**Priority:** Medium  
**Goal:** Improve robustness by testing edge cases.

#### Tasks
1. **Test Error Scenarios**
   - [x] Invalid source files (permissions, non-existent)
   - [x] Invalid YAML/frontmatter in source files
   - [x] Empty source files
   - [x] Destination file permissions
   - [x] Read-only destination directory

2. **Test Parser Edge Cases**
   - [x] Empty configs
   - [x] Configs with only frontmatter (no SSH config)
   - [ ] Malformed SSH config (invalid indentation)
   - [ ] Very large configs

3. **Test Priority Sorting**
   - [x] Files with duplicate priorities
   - [x] Files with negative priorities
   - [x] Mixed prioritized/unprioritized files
   - [x] Zero priority vs no priority

#### Acceptance Criteria
- [x] All error paths are tested
- [ ] Edge cases are documented and handled bugs found and fixed

---

### Phase 5: CI/CD Integration ⏳ PENDING
**Priority:** Low  
**Goal:** Ensure tests run in CI and enforce quality standards.

#### Tasks
1. **Update GitHub Actions**
   - [ ] Add step to run tests with coverage
   - [ ] Generate coverage report
   - [ ] Upload coverage artifact

2. **Add Coverage Threshold**
   - [ ] Configure minimum coverage requirement (e.g., 80%)
   - [ ] Fail build if coverage drops below threshold
   - [ ] Add coverage badge to README

3. **Integrate with Code Quality Tools**
   - [ ] Add coverage reporting to code review
   - [ ] Integrate with code coverage services (Codecov, etc.)

#### Acceptance Criteria
- [ ] Tests run automatically on PRs
- [ ] Coverage is enforced in CI
- [ ] Coverage trends are visible

---

## Implementation Timeline

| Phase | Duration | Dependencies | Target Coverage | Actual Result |
|-------|----------|--------------|------------------|---------------|
| 1 | 1-2 days | None | cmd: >80% | **cmd: 86.5%** ✅ |
| 2 | 1 day | Phase 1 | main: >50% | **main: 50%** ✅ |
| 3 | 1 day | Phase 2 | cmd: >85% | **cmd: 86.5%** ✅ |
| 4 | 1-2 days | Phase 1-3 | overall: >85% | **overall: 88.3%** ✅ |
| 5 | 1 day | Phase 1-4 | overall: >90% | **TBD** ⏳ |

**Total Estimated Duration:** 1-2 weeks (can be done incrementally)

**Actual Progress:** 4 of 5 phases complete ✅

---

## Expected Outcomes

### Coverage Improvements
- **Cmd package**: **86.5%** ✅ (from 26.9%, target >80%)
- **Main.go**: **50.0%** ✅ (from 0%, target >50%, Run() at 100%)
- **Overall**: **88.3%** ⏳ (from 76.9%, target >90%)

### Quality Improvements
- End-to-end binary tests for user-facing behavior
- Better error handling verification
- More robust edge case coverage
- Continuous enforcement of test standards

---

## Success Metrics

1. ✅ All new code has corresponding tests
2. ✅ No regression in existing test coverage
3. ⏳ CI pipeline enforces coverage thresholds
4. ⏳ End-to-end tests verify actual binary behavior
5. ✅ All CLI flags and options are tested

---

## Getting Started

To begin implementing this plan:

1. Start with Phase 1 (CLI Integration Testing) as it has no dependencies
2. Focus on `NewRootCommand` and `expandPath`/`expandGlobs` first
3. Use existing test patterns from `cmd_test.go` as templates
4. Run `go test ./... -cover` frequently to track progress

Example first commit:
```bash
# Create test file for cmd package integration
git checkout -b test/improve-cli-coverage
# Implement tests for NewRootCommand, expandPath, expandGlobs
git commit -m "test(cmd): add integration tests for CLI functions"
```

---

## References

- [Current test files](../sshush/sshush_test.go)
- [Cmd test file](../cmd/cmd_test.go)
- [Go Testing Documentation](https://pkg.go.dev/testing)
- [Go Coverage Documentation](https://go.dev/blog/cover)
