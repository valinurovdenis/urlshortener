package urlstorage

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"

	"github.com/valinurovdenis/urlshortener/internal/app/utils"
)

type URLDump struct {
	UUID        int64  `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type DumpWriter struct {
	file   *os.File
	writer *bufio.Writer
}

func NewDumpWriter(filename string) (*DumpWriter, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &DumpWriter{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (p *DumpWriter) Write(dump URLDump) error {
	data, err := json.Marshal(dump)
	if err != nil {
		return err
	}

	data = append(data, '\n')
	if _, err := p.writer.Write(data); err != nil {
		return err
	}
	return p.writer.Flush()
}

type FileDumpWrapper struct {
	URLStorage
	filename   string
	dumpWriter *DumpWriter
	counter    int64
	dumpMutex  sync.Mutex
}

func (f *FileDumpWrapper) StoreWithContext(ctx context.Context, longURL string, shortURL string, _ string) error {
	if err := f.URLStorage.StoreWithContext(ctx, longURL, shortURL, ""); err != nil {
		return err
	}

	f.dumpMutex.Lock()
	defer f.dumpMutex.Unlock()
	f.counter += 1
	dump := URLDump{UUID: f.counter, ShortURL: shortURL, OriginalURL: longURL}
	return f.dumpWriter.Write(dump)
}

func (f *FileDumpWrapper) RestoreFromDump() error {
	f.URLStorage.Clear()
	var long2ShortUrls []utils.URLPair
	file, err := os.OpenFile(f.filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	data, err := reader.ReadBytes('\n')
	dump := URLDump{}
	for err == nil {
		err = json.Unmarshal(data, &dump)
		if err != nil {
			return err
		}
		url := utils.URLPair{Short: dump.ShortURL, Long: dump.OriginalURL}
		long2ShortUrls = append(long2ShortUrls, url)
		f.counter = dump.UUID
		data, err = reader.ReadBytes('\n')
	}
	if err != io.EOF {
		return err
	}
	f.URLStorage.StoreManyWithContext(context.Background(), long2ShortUrls, "")
	return nil
}

func NewFileDumpWrapper(filename string, storage URLStorage) (*FileDumpWrapper, error) {
	dumpWriter, err := NewDumpWriter(filename)
	if err != nil {
		return nil, err
	}
	return &FileDumpWrapper{
		URLStorage: storage,
		filename:   filename,
		dumpWriter: dumpWriter,
		counter:    0,
	}, nil
}
