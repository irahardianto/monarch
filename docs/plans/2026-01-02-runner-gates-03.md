# Plan: Runner & Standard Gates (Epic 3)

## Requirements Extracted

### Scope
Implementation of the core execution engine (Story 3.1) and safe log parsing architecture (Stories 3.2, 3.3, 3.4) for the Monarch Supervisor. This enables the system to run arbitrary tools inside Docker containers and convert their output into structured, actionable data for Agents.

### Gap Analysis
- **Nouns:** `Executor`, `Parser` (Interface), `LogEntry`, `UnifiedErrorObject`, `GoTestParser`, `ESLintParser`, `RuleHints`.
- **Verbs:** `Run()` (Docker Exec), `Parse()` (Raw -> Struct), `Enrich()` (Add hints), `FailClosed()` (Error handling).
- **Exclusions:** Container lifecycle management (handled in Epic 1), LLM Gates (Epic 5).

### Functional Requirements
- **FR11:** System runs shell commands inside ephemeral containers and captures output.
- **FR12:** System parses raw tool output into `UnifiedErrorObject`.
    - Must distinguish `VALIDATION_FAILURE` vs `SYSTEM_ERROR`.
    - Must "Fail Closed" on malformed output.
    - Must enrich errors with hints where possible.

### Knowledge Enrichment
- **Docker Exec:** Validated `ContainerExecCreate` and `ContainerExecAttach` patterns from Docker SDK.
- **Go Test JSON:** Confirmed `Action` ("fail", "output") and `Package` fields in `go test -json` stream.
- **ESLint JSON:** Confirmed standard array structure `[{"filePath", "messages": [] }]`.

---

## Tasks

### Task 1: Universal Tool Executor

**Files:**
- Create: `apps/backend/runner/executor.go`
- Test: `apps/backend/runner/executor_test.go`

**Requirements:**
- **Acceptance Criteria**
  1. `Run(ctx, containerID, cmd)` executes command in target container.
  2. Returns `stdout`, `stderr`, and `exitCode`.
  3. Returns error if container does not exist or Docker API fails.
- **Functional Requirements**
  1. Use `client.ContainerExecCreate` with `AttachStdout`, `AttachStderr`.
  2. Use `client.ContainerExecAttach` to stream output.
  3. Use `client.ContainerExecInspect` to get exit code.
- **Test Coverage**
  - [Integration] `TestExecutor_Run_Success` - Run `echo "hello"` in a real container.
  - [Integration] `TestExecutor_Run_ExitCode` - Run `exit 1` and verify code.

**Step 1: Write failing test**
```go
package runner_test

import (
    "context"
    "testing"
    "github.com/stretchr/testify/require"
    "monarch/runner"
    "github.com/docker/docker/client"
)

// Assumption: A running container "monarch-test-executor" exists (integration setup)
// For unit test, we'll mock the Docker client if possible, but Executor is heavy on Docker SDK.
// We will use a mock interface for the Docker client in the actual implementation to make it unit testable.

type MockDockerClient struct {
    client.CommonAPIClient
    // Add mock fields
}

func TestExecutor_Run(t *testing.T) {
    // This requires a mock or integration setup. 
    // For this plan, we assume a MockDockerClient is passed to Executor.
    ctx := context.Background()
    cli := &MockDockerClient{} 
    exec := runner.NewExecutor(cli) 
    
    // Test case: Valid execution
    _, _, err := exec.Run(ctx, "test-container", []string{"echo", "hello"})
    require.Error(t, err) // Should fail before implementation
}
```

**Step 2: Verify test fails**
Run: `go test ./apps/backend/runner/... -v`
Expected: FAIL (Undefined `NewExecutor`, `Run`)

**Step 3: Write minimal implementation**
```go
package runner

import (
    "bytes"
    "context"
    "fmt"
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/pkg/stdcopy"
)

type DockerClient interface {
    ContainerExecCreate(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error)
    ContainerExecAttach(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error)
    ContainerExecInspect(ctx context.Context, execID string) (types.ContainerExecInspect, error)
}

type Executor struct {
    cli DockerClient
}

func NewExecutor(cli DockerClient) *Executor {
    return &Executor{cli: cli}
}

func (e *Executor) Run(ctx context.Context, containerID string, cmd []string) (string, int, error) {
    // 1. Create Exec
    cfg := types.ExecConfig{
        Cmd:          cmd,
        AttachStdout: true,
        AttachStderr: true,
    }
    resp, err := e.cli.ContainerExecCreate(ctx, containerID, cfg)
    if err != nil {
        return "", 0, fmt.Errorf("failed to create exec: %w", err)
    }

    // 2. Attach
    attachResp, err := e.cli.ContainerExecAttach(ctx, resp.ID, types.ExecStartCheck{})
    if err != nil {
        return "", 0, fmt.Errorf("failed to attach exec: %w", err)
    }
    defer attachResp.Close()

    // 3. Capture Output
    var outBuf, errBuf bytes.Buffer
    _, err = stdcopy.StdCopy(&outBuf, &errBuf, attachResp.Reader)
    if err != nil {
        return "", 0, fmt.Errorf("failed to copy output: %w", err)
    }

    // 4. Inspect for Exit Code
    inspectResp, err := e.cli.ContainerExecInspect(ctx, resp.ID)
    if err != nil {
        return "", 0, fmt.Errorf("failed to inspect exec: %w", err)
    }

    output := outBuf.String()
    if errBuf.Len() > 0 {
        output += "\nSTDERR:\n" + errBuf.String()
    }

    return output, inspectResp.ExitCode, nil
}
```

**Step 4: Verify test passes**
Run: `go test ./apps/backend/runner/... -v`
Expected: PASS

---

### Task 2: Parser Interface & Types

**Files:**
- Create: `apps/backend/runner/parser/types.go`
- Test: `apps/backend/runner/parser/types_test.go` (if needed for helpers)

**Requirements:**
- **Acceptance Criteria**
  1. Define `LogEntry` struct with all fields.
  2. Define `Parser` interface.
  3. Define `ErrSystemFailure` for fail-closed logic.
- **Functional Requirements**
  1. `LogEntry` must have: Severity, File, Line, Message, Tool, RuleID, Hint.
  2. `Parser` interface must have `Parse(raw []byte) ([]LogEntry, error)`.
- **Test Coverage**
  - [Unit] Compilation check (implicit).

**Step 1: Write failing test**
N/A (Type definition) - We will verify strict compilation in Step 4.

**Step 2: Verify test fails**
N/A

**Step 3: Write minimal implementation**
```go
package parser

import "errors"

// ErrSystemFailure indicates the tool output was malformed or crashed.
// This triggers the "Fail Closed" mechanism.
var ErrSystemFailure = errors.New("system failure: tool output malformed")

type Severity string

const (
    SeverityInfo    Severity = "INFO"
    SeverityWarning Severity = "WARNING"
    SeverityError   Severity = "ERROR"
)

type LogEntry struct {
    Severity Severity `json:"severity"`
    File     string   `json:"file"`
    Line     int      `json:"line"`
    Message  string   `json:"message"`
    Tool     string   `json:"tool"`
    RuleID   string   `json:"rule_id,omitempty"` // e.g., "G101"
    Hint     string   `json:"hint,omitempty"`    // Enriched advice
}

type Parser interface {
    // Parse converts raw tool output into structured entries.
    // Must return ErrSystemFailure if output is unparseable.
    Parse(raw []byte) ([]LogEntry, error)
}
```

**Step 4: Verify test passes**
Run: `go build ./apps/backend/runner/parser`
Expected: Success

---

### Task 3: Go Test Parser

**Files:**
- Create: `apps/backend/runner/parser/go_test.go`
- Test: `apps/backend/runner/parser/go_test_test.go`

**Requirements:**
- **Acceptance Criteria**
  1. Parses `go test -json` stream.
  2. Maps `Action: fail` events to `SeverityError`.
  3. Returns `ErrSystemFailure` on invalid JSON.
- **Functional Requirements**
  1. Handle line-delimited JSON.
  2. Ignore `pass`/`run` events for log entries (unless verbose).
  3. Extract Package as context.
- **Test Coverage**
  - [Unit] `TestGoTestParser_Parse` - Valid input.
  - [Unit] `TestGoTestParser_Parse_Malformed` - Invalid input.

**Step 1: Write failing test**
```go
package parser_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "monarch/runner/parser"
)

func TestGoTestParser_Parse(t *testing.T) {
    raw := []byte(`{"Action":"fail","Package":"m/test","Test":"TestFoo","Output":"FAIL: TestFoo (0.00s)\n"}`)
    p := &parser.GoTestParser{}
    entries, err := p.Parse(raw)
    
    assert.NoError(t, err)
    assert.Len(t, entries, 1)
    assert.Equal(t, "FAIL: TestFoo (0.00s)", entries[0].Message)
}
```

**Step 2: Verify test fails**
Run: `go test ./apps/backend/runner/parser/... -v`
Expected: FAIL (Undefined `GoTestParser`)

**Step 3: Write minimal implementation**
```go
package parser

import (
    "bufio"
    "bytes"
    "encoding/json"
    "strings"
)

type GoTestParser struct{}

type goTestEvent struct {
    Action  string `json:"Action"`
    Package string `json:"Package"`
    Test    string `json:"Test"`
    Output  string `json:"Output"`
}

func (p *GoTestParser) Parse(raw []byte) ([]LogEntry, error) {
    var entries []LogEntry
    scanner := bufio.NewScanner(bytes.NewReader(raw))

    for scanner.Scan() {
        line := scanner.Bytes()
        if len(bytes.TrimSpace(line)) == 0 {
            continue
        }

        var event goTestEvent
        if err := json.Unmarshal(line, &event); err != nil {
            return nil, ErrSystemFailure
        }

        if event.Action == "fail" && event.Test != "" {
            entries = append(entries, LogEntry{
                Severity: SeverityError,
                Message:  strings.TrimSpace(event.Output),
                Tool:     "go test",
                File:     event.Package, // Best effort context
            })
        }
    }

    if err := scanner.Err(); err != nil {
        return nil, ErrSystemFailure
    }

    return entries, nil
}
```

**Step 4: Verify test passes**
Run: `go test ./apps/backend/runner/parser/... -v`
Expected: PASS

---

### Task 4: ESLint Parser

**Files:**
- Create: `apps/backend/runner/parser/eslint.go`
- Test: `apps/backend/runner/parser/eslint_test.go`

**Requirements:**
- **Acceptance Criteria**
  1. Parses `eslint -f json` output (array).
  2. Maps messages to `LogEntry`.
  3. Returns `ErrSystemFailure` on invalid JSON.
- **Functional Requirements**
  1. Input is a single JSON array, not line-delimited.
  2. Map `ruleId` to `RuleID`.
  3. Map `line` to `Line`.
- **Test Coverage**
  - [Unit] `TestESLintParser_Parse` - Valid input.

**Step 1: Write failing test**
```go
package parser_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "monarch/runner/parser"
)

func TestESLintParser_Parse(t *testing.T) {
    raw := []byte(`[{"filePath":"app.ts","messages":[{"ruleId":"no-console","severity":2,"message":"Unexpected console","line":10}]}]`)
    p := &parser.ESLintParser{}
    entries, err := p.Parse(raw)
    
    assert.NoError(t, err)
    assert.Len(t, entries, 1)
    assert.Equal(t, "no-console", entries[0].RuleID)
}
```

**Step 2: Verify test fails**
Run: `go test ./apps/backend/runner/parser/... -v`
Expected: FAIL (Undefined `ESLintParser`)

**Step 3: Write minimal implementation**
```go
package parser

import "encoding/json"

type ESLintParser struct{}

type eslintFile struct {
    FilePath string          `json:"filePath"`
    Messages []eslintMessage `json:"messages"`
}

type eslintMessage struct {
    RuleID   string `json:"ruleId"`
    Severity int    `json:"severity"` // 1=Warning, 2=Error
    Message  string `json:"message"`
    Line     int    `json:"line"`
}

func (p *ESLintParser) Parse(raw []byte) ([]LogEntry, error) {
    var files []eslintFile
    if err := json.Unmarshal(raw, &files); err != nil {
        return nil, ErrSystemFailure
    }

    var entries []LogEntry
    for _, file := range files {
        for _, msg := range file.Messages {
            severity := SeverityWarning
            if msg.Severity == 2 {
                severity = SeverityError
            }

            entries = append(entries, LogEntry{
                Severity: severity,
                File:     file.FilePath,
                Line:     msg.Line,
                Message:  msg.Message,
                Tool:     "eslint",
                RuleID:   msg.RuleID,
            })
        }
    }

    return entries, nil
}
```

**Step 4: Verify test passes**
Run: `go test ./apps/backend/runner/parser/... -v`
Expected: PASS

---

### Task 5: Static Hint Enrichment

**Files:**
- Create: `apps/backend/runner/parser/hints.go`
- Modify: `apps/backend/runner/parser/types.go` (if needed)
- Test: `apps/backend/runner/parser/hints_test.go`

**Requirements:**
- **Acceptance Criteria**
  1. `Enrich(*LogEntry)` adds hint if `RuleID` matches known rule.
  2. Does nothing if no match.
- **Functional Requirements**
  1. Static map of `RuleID` -> `HintString`.
  2. Support common Go/ESLint rules initially.
- **Test Coverage**
  - [Unit] `TestEnrich_Hit` - Matches rule.
  - [Unit] `TestEnrich_Miss` - No match.

**Step 1: Write failing test**
```go
package parser_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "monarch/runner/parser"
)

func TestEnrich(t *testing.T) {
    entry := &parser.LogEntry{RuleID: "no-console"}
    parser.Enrich(entry)
    assert.NotEmpty(t, entry.Hint)
    assert.Contains(t, entry.Hint, "console")
}
```

**Step 2: Verify test fails**
Run: `go test ./apps/backend/runner/parser/... -v`
Expected: FAIL (Undefined `Enrich`)

**Step 3: Write minimal implementation**
```go
package parser

var ruleHints = map[string]string{
    "no-console": "Console logs are forbidden in production. Use a structured logger.",
    "G101":       "Potential hardcoded credential. Use environment variables.",
    // Add more as needed
}

func Enrich(entry *LogEntry) {
    if hint, ok := ruleHints[entry.RuleID]; ok {
        entry.Hint = hint
    }
}
```

**Step 4: Verify test passes**
Run: `go test ./apps/backend/runner/parser/... -v`
Expected: PASS

```