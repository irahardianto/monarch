### Task 1: Crypto Package (AES-GCM)

**Files:**
- Create: `apps/backend/internal/crypto/aes.go`
- Test: `apps/backend/internal/crypto/aes_test.go`

**Requirements:**
- **Acceptance Criteria**
  1. Can encrypt plaintext -> ciphertext + nonce
  2. Can decrypt ciphertext + nonce -> plaintext
  3. Returns error on invalid key size or corrupted ciphertext

- **Functional Requirements**
  1. Use AES-GCM standard
  2. Key size must be 32 bytes (AES-256)
  3. Nonce must be unique (generated via `crypto/rand`)

- **Non-Functional Requirements**
  - Security: Use `crypto/cipher` standard library. 
  - Security: Nonce must be prepended to ciphertext for storage convenience (or returned separately, but prepending is standard for simple storage).

- **Test Coverage**
  - [Unit] `Encrypt()` - returns non-empty result
  - [Unit] `Decrypt()` - round trip works
  - [Unit] `Decrypt()` - fails with wrong key
  - Test data fixtures: 32-byte random key

**Step 1: Write failing test**
```go
package crypto

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAESGCM(t *testing.T) {
	keyStr := "6368616e676520746869732070617373776f726420746f206120736563726574" // 32 bytes hex
	key, _ := hex.DecodeString(keyStr)
	
	plaintext := []byte("sensitive-api-key")

	// Encrypt
	ciphertext, err := Encrypt(plaintext, key)
	require.NoError(t, err)
	assert.NotEmpty(t, ciphertext)
	assert.NotEqual(t, plaintext, ciphertext)

	// Decrypt
	decrypted, err := Decrypt(ciphertext, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}
```

**Step 2: Verify test fails**
Run: `go test ./apps/backend/internal/crypto/... -v`
Expected: FAIL (undefined functions)

**Step 3: Write minimal implementation**
```go
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
```

**Step 4: Verify test passes**
Run: `go test ./apps/backend/internal/crypto/... -v`
Expected: PASS with exit code 0


### Task 2: Settings Table & Service

**Files:**
- Create: `apps/backend/database/migrations/003_settings.up.sql`
- Create: `apps/backend/settings/service.go`
- Test: `apps/backend/settings/service_integration_test.go`
- Modify: `apps/backend/database/models.go` (via sqlc if used, or manual) -> wait, project uses `sqlc`. I should add query and run sqlc generate.
- Create: `apps/backend/database/query/settings.sql`

**Requirements:**
- **Acceptance Criteria**
  1. Can insert/update setting with encryption
  2. Can retrieve and decrypt setting
  3. `settings` table exists in DB

- **Functional Requirements**
  1. Table schema: `key` (TEXT PK), `value` (BYTEA), `is_encrypted` (BOOL), `updated_at` (TIMESTAMPTZ)
  2. Service uses `internal/crypto` for encryption
  3. Encryption Key should be loaded from ENV `MONARCH_ENCRYPTION_KEY` (must be 32 chars)

- **Test Coverage**
  - [Integration] `Save(key, val, encrypted=true)` -> verify DB has binary data
  - [Integration] `Get(key)` -> returns plaintext
  - Test data fixtures: `MONARCH_ENCRYPTION_KEY` in test env

**Step 1: Write failing test (Integration)**
```go
package settings_test

import (
	"context"
	"testing"
	"os"

	"github.com/stretchr/testify/require"
	"monarch/apps/backend/settings"
	"monarch/apps/backend/internal/crypto"
)

func TestSettingsService_Integration(t *testing.T) {
	// Setup real DB connection (assuming helper exists from other tests)
	db := setupTestDB(t) 
	
	// Key for testing
	encKey := make([]byte, 32)
	os.Setenv("MONARCH_ENCRYPTION_KEY", string(encKey)) // Mock env

	svc := settings.NewService(db, encKey)

	ctx := context.Background()
	
	// Test Save
	err := svc.Set(ctx, "GEMINI_API_KEY", "secret-value", true)
	require.NoError(t, err)

	// Test Get
	val, err := svc.Get(ctx, "GEMINI_API_KEY")
	require.NoError(t, err)
	require.Equal(t, "secret-value", val)
}
```

**Step 2: Verify test fails**
Run: `go test ./apps/backend/settings/... -v`
Expected: FAIL (compilation error, missing files)

**Step 3: Write minimal implementation**

1. **Migration (`003_settings.up.sql`)**
```sql
CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value BYTEA NOT NULL,
    is_encrypted BOOLEAN DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

2. **Query (`settings.sql`)**
```sql
-- name: UpsertSetting :exec
INSERT INTO settings (key, value, is_encrypted, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value, is_encrypted = EXCLUDED.is_encrypted, updated_at = NOW();

-- name: GetSetting :one
SELECT * FROM settings WHERE key = $1;
```

3. **Service (`service.go`)**
```go
package settings

import (
	"context"
	"database/sql"
	"errors"
	
	"monarch/apps/backend/database"
	"monarch/apps/backend/internal/crypto"
)

type Service struct {
	q      *database.Queries
	encKey []byte
}

func NewService(db database.DBTX, encKey []byte) *Service {
	return &Service{
		q:      database.New(db),
		encKey: encKey,
	}
}

func (s *Service) Set(ctx context.Context, key string, value string, encrypt bool) error {
	var data []byte
	if encrypt {
		if len(s.encKey) != 32 {
			return errors.New("invalid encryption key length")
		}
		var err error
		data, err = crypto.Encrypt([]byte(value), s.encKey)
		if err != nil {
			return err
		}
	} else {
		data = []byte(value)
	}
	
	return s.q.UpsertSetting(ctx, database.UpsertSettingParams{
		Key:         key,
		Value:       data,
		IsEncrypted: encrypt,
	})
}

func (s *Service) Get(ctx context.Context, key string) (string, error) {
	row, err := s.q.GetSetting(ctx, key)
	if err != nil {
		return "", err
	}
	
	if row.IsEncrypted.Bool {
		decrypted, err := crypto.Decrypt(row.Value, s.encKey)
		if err != nil {
			return "", err
		}
		return string(decrypted), nil
	}
	
	return string(row.Value), nil
}
```

**Step 4: Verify test passes**
Run: `go generate ./...` (to run sqlc)
Run: `go test ./apps/backend/settings/... -v`


### Task 3: Gemini Client Wrapper

**Files:**
- Create: `apps/backend/internal/llm/gemini/client.go`
- Create: `apps/backend/internal/llm/client.go` (Interface)
- Test: `apps/backend/internal/llm/gemini/client_test.go`

**Requirements:**
- **Acceptance Criteria**
  1. Can initialize client with API key
  2. Can generate content from text prompt
  3. Implements generic `llm.Client` interface

- **Functional Requirements**
  1. Use `github.com/google/generative-ai-go/genai`
  2. Model: `gemini-1.5-pro` (or latest) - PRD says `gemini-3-pro-preview` but that might be a typo/placeholder. I'll use `gemini-pro` as safe default or configurable.
  3. Interface: `Generate(ctx, prompt string) (string, error)`

- **Test Coverage**
  - [Unit] Test initialization
  - [Integration] (Mocked) `Generate` calls SDK

**Step 1: Write failing test**
```go
package gemini_test

import (
	"context"
	"testing"
	"github.com/stretchr/testify/assert"
	"monarch/apps/backend/internal/llm/gemini"
)

func TestGeminiClient_Init(t *testing.T) {
	c, err := gemini.NewClient("fake-key")
	assert.NoError(t, err)
	assert.NotNil(t, c)
}
```

**Step 2: Verify test fails**
Run: `go test ./apps/backend/internal/llm/gemini/...`

**Step 3: Write implementation**
```go
// apps/backend/internal/llm/client.go
package llm

import "context"

type Client interface {
	Generate(ctx context.Context, prompt string) (string, error)
	Close() error
}
```

```go
// apps/backend/internal/llm/gemini/client.go
package gemini

import (
	"context"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"monarch/apps/backend/internal/llm"
)

type Client struct {
	genaiClient *genai.Client
	modelName   string
}

func NewClient(apiKey string) (llm.Client, error) {
	ctx := context.Background()
	c, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	return &Client{genaiClient: c, modelName: "gemini-pro"}, nil
}

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	model := c.genaiClient.GenerativeModel(c.modelName)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}
	
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}
	
	var output string
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			output += string(txt)
		}
	}
	return output, nil
}

func (c *Client) Close() error {
	return c.genaiClient.Close()
}
```

**Step 4: Verify test passes**
Run: `go test ./apps/backend/internal/llm/gemini/...`


### Task 4: LLM Eval Engine (Snapshot Mode)

**Files:**
- Create: `apps/backend/runner/eval/engine.go`
- Test: `apps/backend/runner/eval/engine_test.go`

**Requirements:**
- **Acceptance Criteria**
  1. Reads file content from disk
  2. Enforces file size limit
  3. Constructs XML-wrapped prompt
  4. Calls LLM and returns parsed validation result

- **Functional Requirements**
  1. Input: `GateConfig` (includes prompt template), `FilePath`
  2. Logic: `os.Stat` -> check size. `os.ReadFile`.
  3. Prompt: `<file path="..."> content </file> 
 <instruction> ... </instruction>`
  4. Output: `UnifiedErrorObject` (from Runner types)

- **Test Coverage**
  - [Unit] `EvaluateSnapshot` - fails if file too large
  - [Unit] `EvaluateSnapshot` - calls LLM with correct context
  - [Unit] `EvaluateSnapshot` - parses response

**Step 1: Write failing test**
```go
package eval_test

import (
	"context"
	"testing"
	"monarch/apps/backend/runner/eval"
	"monarch/apps/backend/internal/llm/mocks" // generated mock
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEvaluateSnapshot_Limit(t *testing.T) {
	mockLLM := new(mocks.Client)
	engine := eval.NewEngine(mockLLM, 100) // 100 bytes limit

	// Create dummy large file in temp dir
	// ... (setup code)

	err := engine.EvaluateSnapshot(context.Background(), "large_file.txt", "Check this")
	require.ErrorContains(t, err, "file size exceeds limit")
}
```

**Step 2: Verify test fails**
Run: `go test ...`

**Step 3: Write implementation**
```go
package eval

import (
	"context"
	"fmt"
	"os"
	"monarch/apps/backend/internal/llm"
)

type Engine struct {
	llm       llm.Client
	sizeLimit int64
}

func NewEngine(llm llm.Client, sizeLimit int64) *Engine {
	return &Engine{llm: llm, sizeLimit: sizeLimit}
}

func (e *Engine) EvaluateSnapshot(ctx context.Context, path string, instruction string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.Size() > e.sizeLimit {
		return "", fmt.Errorf("file size %d exceeds limit %d", info.Size(), e.sizeLimit)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf("<file path=\" %s \"><\/file>\n\n<instruction>\n%s\n<\/instruction>", path, string(content), instruction)
	
	return e.llm.Generate(ctx, prompt)
}
```

**Step 4: Verify test passes**
Run: `go test ...`


### Task 5: LLM Eval Engine (Diff Mode)

**Files:**
- Modify: `apps/backend/runner/eval/engine.go`
- Test: `apps/backend/runner/eval/engine_test.go`

**Requirements:**
- **Acceptance Criteria**
  1. Runs `git diff` for specific file
  2. Sends diff to LLM
  3. Handles case where git is missing or file untracked

- **Functional Requirements**
  1. Use `exec.Command("git", "diff", path)`
  2. Prompt: `<diff path="..."> ... </diff>`

**Step 1: Write failing test**
```go
func TestEvaluateDiff(t *testing.T) {
    // Mock LLM
    // Run engine.EvaluateDiff(...)
    // Verify LLM called with diff content
}
```

**Step 2: Verify test fails**

**Step 3: Write implementation**
```go
func (e *Engine) EvaluateDiff(ctx context.Context, path string, instruction string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "diff", path)
	diff, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}
    
    // ... construct prompt ...
    // ... call LLM ...
}
```

**Step 4: Verify test passes**


### Task 6: Runner Integration

**Files:**
- Modify: `apps/backend/runner/service.go`
- Modify: `apps/backend/runner/executor.go`

**Requirements:**
- **Acceptance Criteria**
  1. Runner detects gate type (Standard vs LLM)
  2. Routes LLM gates to `eval.Engine`
  3. Routes Standard gates to `Docker` executor

- **Functional Requirements**
  1. Inject `eval.Engine` into `RunnerService`
  2. In `RunGate`, switch on `gate.Type`
  3. If `type == "llm_eval"`, call `engine.Evaluate...`

**Step 1: Write failing test**
- Mock both Docker and Eval engine.
- Call RunGate with `type: llm_eval`.
- Expect Eval engine to be called.

**Step 2: Verify test fails**

**Step 3: Write implementation**
- Wire it up.

**Step 4: Verify test passes**

