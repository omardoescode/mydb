package wal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
)

/**
* NOTE: WAL Requirements / Design
* 1. Support for Undo + Redo
* 2. Append(Body) -> LSN
* 3. Support for checkpoint as a bg task
 */

/**
* TODO:
* - Periodic flushing into btree data pages
* - Crash Recovery with checkpoints
* - this interface isn't concurrent-compatiable yet
* - Append must be synchronous on the `Append` method:
* 	- The WAL protocol asserts that the log records representing changes to some data must already be on stable storage before the changed data is allowed to replace the previous version of that data on nonvolatile storage.
* 	- Append's fsync-every-call is overkill but correct; later optimize to fsync only on Commit record + group commit
 */

const WAL_DIRECTORY_PATH = "./mydb_wal/"

type WAL struct {
	nextLSN    LSN
	flushedLSN LSN
	appliedLSN LSN
	file       *os.File
}

func New(filePath string) (*WAL, error) {
	err := os.MkdirAll(WAL_DIRECTORY_PATH, 0755)
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path.Join(WAL_DIRECTORY_PATH, filePath), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	return &WAL{
		file:       file,
		nextLSN:    1,
		flushedLSN: 0,
		appliedLSN: 0,
	}, nil
}

// TODO: Change thsi signature: Body should be an internal struct
func (w *WAL) Append(b Body) (LSN, error) {
	lsn := w.nextLSN
	if _, err := writeEntry(w.file, lsn, b); err != nil {
		return 0, err
	}
	if err := w.file.Sync(); err != nil {
		return 0, err
	}
	w.nextLSN++
	w.flushedLSN = lsn
	return lsn, nil
}

func Restore(filePath string) (*WAL, error) {
	_, err := os.Stat(path.Join(WAL_DIRECTORY_PATH, filePath))
	if err != nil {
		return nil, err
	}
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("WAL: file \"%s\" doesn't exist", filePath)
	}

	// Read the entire file
	file, err := os.OpenFile(
		path.Join(WAL_DIRECTORY_PATH, filePath), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)

	if err != nil {
		return nil, err
	}

	var nextLSN LSN = 1
	var flushedLSN LSN = 0

	wal := &WAL{
		file:       file,
		nextLSN:    nextLSN,
		flushedLSN: LSN(flushedLSN),
		appliedLSN: 0,
	}

	for {
		entry, err := readEntry(file)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			if errors.Is(err, io.ErrUnexpectedEOF) {
				return nil, err
			}
		}

		nextLSN++
		// NOTE: Here we assume all have been flushed
		flushedLSN++

		fmt.Println(entry)
	}

	return wal, nil
}
