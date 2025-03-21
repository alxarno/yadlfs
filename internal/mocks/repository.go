//nolint:wrapcheck,forcetypeassert
package mocks

import (
	"context"
	"errors"
	"io"

	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
	BufferSize int64
}

func (m *MockRepository) Upload(ctx context.Context, filePath string, r io.Reader, overwrite bool) error {
	args := m.Called(ctx, filePath, r, overwrite)
	b := make([]byte, m.BufferSize)

	for {
		_, err := r.Read(b)
		if errors.Is(err, io.EOF) {
			break
		}
	}

	return args.Error(0)
}

func (m *MockRepository) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	args := m.Called(ctx, path)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(io.ReadCloser), args.Error(1)
}
