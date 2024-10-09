/**
 * This software and associated documentation files (the “Software”),
 * including GFI AppManager, is the property of GFI USA, LLC and its affiliates.
 * No part of the Software may be copied, modified, distributed, sold, or otherwise
 * used except as expressly permitted by the terms of the software license agreement.
 */

package logger

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

var (
	fullIndexFilename string
	fullLogFilename   string
	offsetData        int64
)

func isEmptyFile(filename string) bool {
	fi, err := os.Stat(filename)
	if err != nil {
		return false
	}

	size := fi.Size()
	return size == 0
}

func updateIndexFile() {
	f, err := os.Open(fullLogFilename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	defer f.Close()
	_, err = f.Seek(offsetData, io.SeekStart)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	indexFile, err := os.OpenFile(fullIndexFilename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer indexFile.Close()

	if isEmptyFile(fullIndexFilename) {
		binary.Write(indexFile, binary.LittleEndian, int64(1))
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(scanEndedLines)

	var offset int64 = 0
	for scanner.Scan() {
		b := scanner.Bytes()
		offset += int64(len(b))
		binary.Write(indexFile, binary.LittleEndian, offsetData + offset)
	}

	offsetData += offset
}

// read last offset and remember it in new "session"
func readLastOffset() {
	f, err := os.Open(fullIndexFilename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	var res int64
	f.Seek(-8, 2)
	err = binary.Read(f, binary.LittleEndian, &res)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	offsetData = res
}

// create log index file if one not exist
func createLogIndexFile() {
	if _, err := os.Stat(LogDir()); os.IsNotExist(err) {
		err = os.Mkdir(LogDir(), 0755)

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}

	indexFile, err := os.OpenFile(fullIndexFilename, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer indexFile.Close()

	logReader, err := os.Open(fullLogFilename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	defer logReader.Close()

	binary.Write(indexFile, binary.LittleEndian, int64(1))

	scanner := bufio.NewScanner(logReader)
	scanner.Split(scanLines)

	var offset int64 = 0

	for scanner.Scan() {
		b := scanner.Bytes()
		offset += int64(len(b))
		binary.Write(indexFile, binary.LittleEndian, offset)
	}

	offsetData = offset

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}

// helper function for scanner
func scanLinesEx(data []byte, atEOF bool, allAtEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		return i + 1, data[0 : i+1], nil
	}

	// If we're at EOF, return all data.
	if atEOF && allAtEOF {
		return len(data), data, nil
	}

	// Request more data.
	return 0, nil, nil
}

func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	return scanLinesEx(data, atEOF, true)
}

func scanEndedLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	return scanLinesEx(data, atEOF, false)
}
