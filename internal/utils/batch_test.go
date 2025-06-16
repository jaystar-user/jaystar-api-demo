package utils

import (
	"errors"
	"github.com/stretchr/testify/mock"
	"testing"
)

type fakeCall interface {
	Do(batchIndex int, start int, end int) error
}
type fakeObject struct {
	mock.Mock
}

func (f *fakeObject) Do(batchIndex int, start int, end int) error {
	args := f.Called(batchIndex, start, end)
	return args.Error(0)
}

func TestRunInBatch(t *testing.T) {
	obj := &fakeObject{}
	obj.On("Do", 1, 0, 2).Return(nil)
	obj.On("Do", 2, 2, 4).Return(nil)
	obj.On("Do", 3, 4, 6).Return(nil)
	obj.On("Do", 4, 6, 8).Return(nil)
	obj.On("Do", 5, 8, 10).Return(nil)
	t.Run("test batch op", func(t *testing.T) {
		err := RunInBatch(10, 2, obj.Do)
		if err != nil {
			t.Errorf("RunInBatch test failed: unexpected error occurred: %v", err)
		}
		obj.AssertExpectations(t)
		obj.AssertNumberOfCalls(t, "Do", 5)
	})

	unexpectedErr := errors.New("error occurred")
	obj = &fakeObject{}
	obj.On("Do", 1, 0, 2).Return(unexpectedErr)
	t.Run("test batch op with error", func(t *testing.T) {
		err := RunInBatch(10, 2, obj.Do)
		if err != nil && !errors.Is(err, unexpectedErr) {
			t.Errorf("RunInBatch test failed: expected return unexpected error but got: %v", err)
		}
		obj.AssertExpectations(t)
		obj.AssertNumberOfCalls(t, "Do", 1)
	})

	obj = &fakeObject{}
	obj.On("Do", 1, 0, 3).Return(nil)
	obj.On("Do", 2, 3, 6).Return(nil)
	obj.On("Do", 3, 6, 9).Return(nil)
	obj.On("Do", 4, 9, 10).Return(nil)
	t.Run("test batch op with non-divisible batchSize", func(t *testing.T) {
		err := RunInBatch(10, 3, obj.Do)
		if err != nil {
			t.Errorf("RunInBatch test failed: unexpected error occurred: %v", err)
		}
		obj.AssertExpectations(t)
		obj.AssertNumberOfCalls(t, "Do", 4)
	})

	t.Run("test batch op with invalid input", func(t *testing.T) {
		err := RunInBatch(10, 0, func(batchIndex int, start int, end int) error {
			return nil
		})
		if err != nil && err.Error() != "invalid batchSize" {
			t.Errorf("RunInBatch test failed: expected error to be invalid batchSize but got: %v", err)
		}
		obj.AssertNotCalled(t, "Do")
	})

	obj = &fakeObject{}
	obj.On("Do", 1, 0, 2).Return(nil)
	t.Run("test batch op with zero total", func(t *testing.T) {
		err := RunInBatch(0, 2, obj.Do)
		if err != nil {
			t.Errorf("RunInBatch test failed: unexpected error occurred: %v", err)
		}
		obj.AssertNotCalled(t, "Do")
	})
}
