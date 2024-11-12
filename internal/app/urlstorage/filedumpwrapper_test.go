package urlstorage

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/valinurovdenis/urlshortener/internal/app/mocks"
)

type Consumer struct {
	file    *os.File
	scanner *bufio.Scanner
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}

func (c *Consumer) ReadDump() (*URLDump, error) {
	if !c.scanner.Scan() {
		return nil, c.scanner.Err()
	}
	data := c.scanner.Bytes()

	dump := URLDump{}
	err := json.Unmarshal(data, &dump)
	if err != nil {
		return nil, err
	}

	return &dump, nil
}

func (c *Consumer) Close() {
	c.file.Close()
}

func (c *Consumer) Rewind() {
	c.file.Seek(0, io.SeekStart)
}

func TestFileDumpWrapper_testDump(t *testing.T) {
	testFilename := "test_dump"
	defer os.Remove(testFilename)
	mockStorage := mocks.NewURLStorage(t)
	mockStorage.On("StoreWithContext", mock.Anything, "http://youtube.ru/1", "1").Return(nil).Once()
	mockStorage.On("StoreWithContext", mock.Anything, "http://youtube.ru/2", "2").Return(nil).Once()
	{
		dumpWrapper, _ := NewFileDumpWrapper(testFilename, mockStorage)

		dumpWrapper.StoreWithContext(context.Background(), "http://youtube.ru/1", "1")
		dumpWrapper.StoreWithContext(context.Background(), "http://youtube.ru/2", "2")
	}

	consumer, _ := NewConsumer(testFilename)
	defer consumer.Close()

	checkEqualDumps := func(num int) {
		consumer.Rewind()
		for i := 1; i < num+1; i++ {
			dump, err := consumer.ReadDump()
			require.Equal(t, nil, err)
			expectedDump := URLDump{
				UUID:        int64(i),
				OriginalURL: "http://youtube.ru/" + strconv.Itoa(i),
				ShortURL:    strconv.Itoa(i)}
			assert.Equal(t, expectedDump, *dump)
		}
	}
	checkEqualDumps(2)

	mockStorage.On("Clear").Return(nil).Once()
	mockStorage.On("StoreManyWithContext", mock.Anything, map[string]string{
		"http://youtube.ru/1": "1",
		"http://youtube.ru/2": "2"}).Return([]error{nil, nil}, nil).Once()
	mockStorage.On("StoreWithContext", mock.Anything, "http://youtube.ru/3", "3").Return(nil).Once()
	{
		dumpWrapper, _ := NewFileDumpWrapper(testFilename, mockStorage)

		dumpWrapper.RestoreFromDump()
		dumpWrapper.StoreWithContext(context.Background(), "http://youtube.ru/3", "3")
	}
	checkEqualDumps(3)
}
