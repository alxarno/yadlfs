package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/sync/semaphore"
)

var (
	ErrUnsupportedOperation = errors.New("unsupported operation")
	ErrOpenFile             = errors.New("failed to open file")
	ErrUploadFailed         = errors.New("upload failed")
	ErrCreateFile           = errors.New("failed to create file")
	ErrDownloadFailed       = errors.New("download failed")
	ErrCopyData             = errors.New("failed to copy data")
	ErrSemaphoreAcquire     = errors.New("failed to acquire semaphore")
)

type repository interface {
	Upload(ctx context.Context, filePath string, r io.Reader, overwrite bool) error
	Download(ctx context.Context, path string) (io.ReadCloser, error)
}

type Controller struct {
	messages  chan DialMessage
	operation OperationName
	semaphore *semaphore.Weighted
	warehouse repository
	folder    string
}

func NewController(warehouse repository, folder string, messages chan DialMessage) *Controller {
	return &Controller{
		messages:  messages,
		semaphore: semaphore.NewWeighted(1),
		warehouse: warehouse,
		folder:    folder,
	}
}

func (s *Controller) upload(ctx context.Context, event Transfer) error {
	f, err := os.Open(event.Path)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOpenFile, err)
	}

	countingReader := newUploadFileProgress(f, event, s.messages)

	if err = s.warehouse.Upload(ctx, event.OID, countingReader, true); err != nil {
		return fmt.Errorf("%w: %w", ErrUploadFailed, err)
	}

	return nil
}

func (s *Controller) download(ctx context.Context, event Transfer) error {
	path := filepath.Join(s.folder, event.OID)
	fileMode := 665

	outputFile, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND, os.FileMode(fileMode))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCreateFile, err)
	}

	downloadReader, err := s.warehouse.Download(ctx, event.OID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDownloadFailed, err)
	}

	countingWriter := newDownloadFileProgress(outputFile, event, s.messages)

	_, err = io.Copy(countingWriter, downloadReader)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCopyData, err)
	}

	return nil
}

func (s *Controller) init(m Init) error {
	s.operation = m.Operation
	s.semaphore = semaphore.NewWeighted(max(m.ConcurrentTransfers, 1))
	s.messages <- ConfirmMessage{}

	return nil
}

func (s *Controller) transfer(ctx context.Context, event Transfer) error {
	if err := s.semaphore.Acquire(ctx, 1); err != nil {
		return fmt.Errorf("%w: %w", ErrSemaphoreAcquire, err)
	}

	go s.handleTransfer(ctx, event)

	return nil
}

func (s *Controller) handleTransfer(ctx context.Context, event Transfer) {
	defer s.semaphore.Release(1)

	var err error

	switch s.operation {
	case OperationNameDownload:
		err = s.download(ctx, event)
	case OperationNameUpload:
		err = s.upload(ctx, event)
	default:
		err = fmt.Errorf("%w: %s", ErrUnsupportedOperation, s.operation)
	}

	if err != nil {
		s.sendErrorMessage(event.OID, err)

		return
	}

	s.sendCompletionMessage(event.OID)
}

func (s *Controller) sendCompletionMessage(oid string) {
	s.messages <- CompleteMessage{OID: oid}
}

func (s *Controller) sendErrorMessage(oid string, err error) {
	s.messages <- CompleteErrorMessage{
		OID:   oid,
		Error: CompleteErrorMessageContent{0, err.Error()},
	}
}
