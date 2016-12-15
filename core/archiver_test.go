package core

import (
	"os"
	"testing"
)

var (
	fileList chan *ArchiveFile
)

func TestSuite(t *testing.T) {
	CmdPrm.Source = "./"
	CmdPrm.FilePattern = "testfile"

	t.Run("Skal finne fil.", shouldFindFile)
	t.Run("Skal ikke finne fil", shouldNotFindFile)
}

func shouldFindFile(t *testing.T) {
	setUpArchiver()

	testFile := "testfile"
	f, _ := os.OpenFile(testFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	f.Close()

	FindFiles(fileList)

	if testFileFound := <-fileList; testFileFound == nil {
		t.Fail()
	}

	os.Remove(testFile)
}

func shouldNotFindFile(t *testing.T) {
	setUpArchiver()
	FindFiles(fileList)

	if testFileFound := <-fileList; testFileFound != nil {
		t.Fail()
	}
}

func setUpArchiver() {
	fileList = make(chan *ArchiveFile, 1)
	StartAsync(1)
}
