# Test Improvement Plan for sshush

This document outlines a detailed plan to improve test coverage and verification for the sshush program. The current test suite has strong functional testing for the core logic but significant gaps in CLI integration and entry point verification.

---

## Current State

### Test Coverage Summary
- **cmd package**: 26.9% coverage
- **sshush package**: 87.8% coverage
- **main.go**: 0% coverage (not tested)
- **Overall**: 76.9% of statements

### Strengths
1. Comprehensive golden file tests for SSH configuration generation
2. Core parser functions are well-tested (82-100% coverage for most functions)
3. Error cases are tested (missing files, bad config)
4. Special modes are verified (dry-run, backup creation, priority ordering)

### Gaps
1. No tests for `main.go` or the program entry point
2. `NewRootCommand` and CLI integration are untested
3. No end-to-end binary execution tests
4. Missing edge cases for path expansion and glob handling

---

## Detailed Plan

### Phase 1: CLI Integration Testing
**Priority:** High  
**Goal:** Achieve >80% coverage for `cmd/cmd.go` and verify end-to-end program execution.

#### Tasks
1. **Test `NewRootCommand`**
   - [ ] Create test that initializes the root command
   - [ ] Verify default flag values (source, dest, verbose, debug, dry-run)
   - [ ] Test command metadata (Use, Short, Version)
   - [ ] Mock `os.UserHomeDir()` to avoid filesystem dependency
   - [ ] Test viper configuration binding

2. **Test `expandPath` Edge Cases**
   - [ ] Tilde expansion with subdirectories (`~/dir/subdir`)
   - [ ] Multiple environment variables (`$HOME/$USER/file`)
   - [ ] Invalid paths (return errors)
   - [ ] Empty string input

3. **Test `expandGlobs` Edge Cases**
   - [ ] Nested glob patterns (`dir/**/*.yml`)
   - [ ] No matches (empty slice)
   - [ ] Invalid glob patterns (error handling)
   - [ ] Mixed valid/invalid patterns

4. **Test `must()` Function**
   - [ ] Verify it calls `os.Exit(1)` on error
   - [ ] Test with various error types

#### Acceptance Criteria
- [ ] `cmd/cmd.go` coverage > 80%
- [ ] All public functions in `cmd` package have at least one test

---

### Phase 2: Main Entry Point Testing
**Priority:** High  
**Goal:** Test `main.go` and achieve >50% coverage.

#### Tasks
1. **Refactor `main()` for Testability**
   - [ ] Extract root command execution logic into a separate `Run()` function
   - [ ] Make logger configurable for testing
   - [ ] Avoid direct `os.Exit()` calls in testable code

2. **Test `Run()` Function**
   - [ ] Mock `os.Stdout` to capture logs
   - [ ] Test successful execution (no error)
   - [ ] Test error handling (verify exit code 1)
   - [ ] Verify logging output contains expected messages

#### Acceptance Criteria
- [ ] `main.go` coverage > 50%
- [ ] Entry point can be tested without building binary

---

### Phase 3: End-to-End Binary Testing
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

### Phase 4: Error Handling and Edge Cases
**Priority:** Medium  
**Goal:** Improve robustness by testing edge cases.

#### Tasks
1. **Test Error Scenarios**
   - [ ] Invalid source files (permissions, non-existent)
   - [ ] Invalid YAML/frontmatter in source files
   - [ ] Empty source files
   - [ ] Destination file permissions
   - [ ] Read-only destination directory

2. **Test Parser Edge Cases**
   - [ ] Empty configs
   - [ ] Configs with only frontmatter (no SSH config)
   - [ ] Malformed SSH config (invalid indentation)
   - [ ] Very large configs

3. **Test Priority Sorting**
   - [ ] Files with duplicate priorities
   - [ ] Files with negative priorities
   - [ ] Mixed prioritized/unprioritized files
   - [ ] Zero priority vs no priority

#### Acceptance Criteria
- [ ] All error paths are tested
- [ ] Edge cases are documented and handled

---

### Phase 5: CI/CD Integration
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

| Phase | Duration | Dependencies | Target Coverage |
|-------|----------|--------------|------------------|
| 1 | 1-2 days | None | cmd: >80% |
| 2 | 1 day | Phase 1 | main: >50% |
| 3 | 1 day | Phase 2 | cmd: >85% |
| 4 | 1-2 days | Phase 1-3 | overall: >85% |
| 5 | 1 day | Phase 1-4 | overall: >90% |

**Total Estimated Duration:** 1-2 weeks (can be done incrementally)

---

## Expected Outcomes

### Coverage Improvements
- **Cmd package**: >80% (from 26.9%)
- **Main.go**: >50% (from 0%)
- **Overall**: >90% (from 76.9%)

### Quality Improvements
- End-to-end binary tests for user-facing behavior
- Better error handling verification
- More robust edge case coverage
- Continuous enforcement of test standards

---

## Success Metrics

1. ✅ All new code has corresponding tests
2. ✅ No regression in existing test coverage
3. ✅ CI pipeline enforces coverage thresholds
4. ✅ End-to-end tests verify actual binary behavior
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
