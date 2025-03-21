//nolint:lll
package internal

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/alxarno/yadlfs/internal/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type taskFile struct {
	OID              string
	Size             int64
	Path             string
	ExpectedProgress []string // Expected progress messages
	ShouldFail       bool     // Whether the upload should fail
}

func (tf taskFile) Command(event string) string {
	return fmt.Sprintf(`{ "event": "%s", "oid": "%s", "size": %d, "path": "%s" }`, event, tf.OID, tf.Size, tf.Path)
}

// Helper function to create test files in a temporary directory.
func createTestFiles(t *testing.T, task taskFile) {
	t.Helper()

	if !task.ShouldFail { // Only create files for tasks that should not fail
		f, err := os.Create(task.Path)
		require.NoError(t, err, "os.Create")
		err = f.Truncate(task.Size)
		require.NoError(t, err, "f.Truncate")
	}
}

// Helper function to set up the test environment (pipes, scanner, etc.)
func setupTestEnvironment(t *testing.T, mockRepo *mocks.MockRepository, tempDir string) (io.Reader, io.Writer, *sync.WaitGroup) {
	t.Helper()

	responses := make(chan DialMessage)
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	wg := &sync.WaitGroup{}

	dial := NewDial(outW, responses)
	controller := NewController(mockRepo, tempDir, responses)
	dispatcher := NewDispatcher(inR, controller)

	ctx, cancelFunc := context.WithCancel(t.Context())
	go dial.ListenAndServe(ctx)

	wg.Add(1)

	go func() {
		defer cancelFunc()
		defer wg.Done()

		err := dispatcher.ListenAndServe(ctx)
		//nolint:testifylint
		require.ErrorIs(t, err, io.EOF)
	}()

	return outR, inW, wg
}

func TestUploadFailed(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	task := taskFile{
		OID:  "invalid_file",
		Size: 1024,
		Path: "nonexistent_file.bin",
		ExpectedProgress: []string{
			`{"event":"complete","oid":"invalid_file","error":{"code":0,"message":"failed to open file: open nonexistent_file.bin: no such file or directory"}}`,
		},
		ShouldFail: true, // Simulate a failure due to a missing file
	}

	// Create test files in a temporary directory
	createTestFiles(t, task)

	// Set up the test environment
	mockRepo := &mocks.MockRepository{BufferSize: 1024}
	outR, inW, wg := setupTestEnvironment(t, mockRepo, tempDir)

	// Send the init message
	fmt.Fprintln(inW, `{ "event": "init", "operation": "upload", "remote": "origin", "concurrent": true, "concurrenttransfers": 1 }`)

	// Verify the init response
	scanner := bufio.NewScanner(outR)
	scanner.Scan()
	require.Equal(t, "{ }", scanner.Text())

	mockRepo.On("Upload", mock.Anything, task.OID, mock.Anything, true).Return(nil)

	fmt.Fprintln(inW, task.Command("upload"))

	for _, expected := range task.ExpectedProgress {
		scanner.Scan()
		require.Equal(t, expected, scanner.Text())
	}

	// Send the terminate message
	fmt.Fprintln(inW, `{ "event": "terminate" }`)

	// Wait for the dispatcher to finish
	wg.Wait()
}

func TestUpload(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	task := taskFile{
		OID:  "0a5070",
		Size: 4096,
		Path: filepath.Join(tempDir, "file.bin"),
		ExpectedProgress: []string{
			`{"event":"progress","oid":"0a5070","bytesSoFar":0,"bytesSinceLast":0}`,
			`{"event":"progress","oid":"0a5070","bytesSoFar":1024,"bytesSinceLast":1024}`,
			`{"event":"progress","oid":"0a5070","bytesSoFar":2048,"bytesSinceLast":1024}`,
			`{"event":"progress","oid":"0a5070","bytesSoFar":3072,"bytesSinceLast":1024}`,
			`{"event":"progress","oid":"0a5070","bytesSoFar":4096,"bytesSinceLast":1024}`,
			`{"event":"complete","oid":"0a5070"}`,
		},
	}

	// Create test files in a temporary directory
	createTestFiles(t, task)

	// Set up the test environment
	mockRepo := &mocks.MockRepository{BufferSize: 1024}
	outR, inW, wg := setupTestEnvironment(t, mockRepo, tempDir)

	// Send the init message
	fmt.Fprintln(inW, `{ "event": "init", "operation": "upload", "remote": "origin", "concurrent": true, "concurrenttransfers": 1 }`)

	// Verify the init response
	scanner := bufio.NewScanner(outR)
	scanner.Scan()
	require.Equal(t, "{ }", scanner.Text())

	mockRepo.On("Upload", mock.Anything, task.OID, mock.Anything, true).Return(nil)

	fmt.Fprintln(inW, task.Command("upload"))

	for _, expected := range task.ExpectedProgress {
		scanner.Scan()
		require.Equal(t, expected, scanner.Text())
	}

	// Send the terminate message
	fmt.Fprintln(inW, `{ "event": "terminate" }`)

	// Wait for the dispatcher to finish
	wg.Wait()
}

func TestDownloadFailed(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	task := taskFile{
		OID:  "invalid_file",
		Size: 1024,
		Path: "nonexistent_file.bin",
		ExpectedProgress: []string{
			`{"event":"complete","oid":"invalid_file","error":{"code":0,"message":"download failed: mock produced error by download method"}}`,
		},
		ShouldFail: true, // Simulate a failure due to a missing file
	}

	// Create test files in a temporary directory
	createTestFiles(t, task)

	// Set up the test environment
	mockRepo := &mocks.MockRepository{BufferSize: 1024}
	outR, inW, wg := setupTestEnvironment(t, mockRepo, tempDir)

	// Send the init message
	fmt.Fprintln(inW, `{ "event": "init", "operation": "download", "remote": "origin", "concurrent": true, "concurrenttransfers": 1 }`)

	// Verify the init response
	scanner := bufio.NewScanner(outR)
	scanner.Scan()
	require.Equal(t, "{ }", scanner.Text())

	//nolint:err113
	mockRepo.On("Download", mock.Anything, task.OID).Return(nil, errors.New("mock produced error by download method"))

	fmt.Fprintln(inW, task.Command("download"))

	for _, expected := range task.ExpectedProgress {
		scanner.Scan()
		require.Equal(t, expected, scanner.Text())
	}

	// Send the terminate message
	fmt.Fprintln(inW, `{ "event": "terminate" }`)

	// Wait for the dispatcher to finish
	wg.Wait()
}

func TestDownload(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	task := taskFile{
		OID:  "0a5070",
		Size: 4096,
		Path: filepath.Join(tempDir, "file.bin"),
		ExpectedProgress: []string{
			`{"event":"progress","oid":"0a5070","bytesSoFar":0,"bytesSinceLast":0}`,
			`{"event":"complete","oid":"0a5070"}`,
		},
	}

	// Create test files in a temporary directory
	createTestFiles(t, task)

	// Set up the test environment
	mockRepo := &mocks.MockRepository{BufferSize: 1024}
	outR, inW, wg := setupTestEnvironment(t, mockRepo, tempDir)

	// Send the init message
	fmt.Fprintln(inW, `{ "event": "init", "operation": "download", "remote": "origin", "concurrent": true, "concurrenttransfers": 1 }`)

	// Verify the init response
	scanner := bufio.NewScanner(outR)
	scanner.Scan()
	require.Equal(t, "{ }", scanner.Text())

	mockRepo.On("Download", mock.Anything, task.OID).Return(io.NopCloser(&io.LimitedReader{R: &io.LimitedReader{}, N: task.Size}), nil)

	fmt.Fprintln(inW, task.Command("download"))

	for _, expected := range task.ExpectedProgress {
		scanner.Scan()
		require.Equal(t, expected, scanner.Text())
	}

	// Send the terminate message
	fmt.Fprintln(inW, `{ "event": "terminate" }`)

	// Wait for the dispatcher to finish
	wg.Wait()

	f, err := os.Open(task.Path)
	require.NoError(t, err)

	fi, err := f.Stat()
	require.NoError(t, err)
	require.Equal(t, task.Size, fi.Size())
}
