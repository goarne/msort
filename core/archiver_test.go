package core

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"
)

var (
	testDirectory              = "./testfolder/"
	testFileName        string = "testfilename"
	ArchiverFileChannel chan *ArchiveFile
)

func TestSuite(t *testing.T) {
	setCmmdPrmForArchiverUnderTest(testFileName)

	createTestFile(testDirectory, testFileName)
	defer os.RemoveAll(testDirectory)

	setUpArchiver()
	t.Run("Should find file.", shouldFindFile)

	setUpArchiver()
	t.Run("Should not find file", shouldNotFindFile)

	setUpArchiver()
	t.Run("Shall archive file", shouldArchiveFile)
}

func shouldArchiveFile(t *testing.T) {
	af := &ArchiveFile{}
	af.FileInfo, _ = os.Stat(testDirectory + testFileName)
	af.SourcePath = CmdPrm.Source + testFileName

	go func() {
		ArchiverFileChannel <- af
		defer close(ArchiverFileChannel)
	}()

	ArchiveFiles(ArchiverFileChannel)

	if fn, _ := archivedFileExists(testFileName); fn == nil {
		println("target:", fn.Name())
		t.Fail()
	}

	cleanupFile(af.TargetPath+af.Name(), t)
}

func shouldFindFile(t *testing.T) {

	FindFiles(ArchiverFileChannel)

	if testFileFound := <-ArchiverFileChannel; testFileFound == nil {
		t.Fail()
	}
}

func shouldNotFindFile(t *testing.T) {
	CmdPrm.FilePattern = "patternNotMatchingTestfile"

	FindFiles(ArchiverFileChannel)

	if testFileFound := <-ArchiverFileChannel; testFileFound != nil {
		t.Fail()
	}
}

//Below are helper functions for setting up unittest and verifying results.
func setCmmdPrmForArchiverUnderTest(fn string) {
	CmdPrm.Source = testDirectory
	CmdPrm.Target = testDirectory
	CmdPrm.FilePattern = fn
	CmdPrm.TargetPattern = "YYYY/MM/DD"
	CmdPrm.ShallArchive = true
	CmdPrm.Overwrite = false
}

func archivedFileExists(fn string) (os.FileInfo, error) {

	year := strconv.Itoa(time.Now().Year())
	month := strconv.Itoa(int(time.Now().Month()))
	day := strconv.Itoa(time.Now().Day())

	return os.Stat(testDirectory + year + "/" + month + "/" + day + "/" + fn)
}

func setUpArchiver() {
	ArchiverFileChannel = make(chan *ArchiveFile, 1)
	StartAsync(1)
}

func createTestFile(path string, fn string) {
	if _, err := os.Stat(path); err != nil {
		os.MkdirAll(path, 0777)
	}

	f, err := os.OpenFile(path+fn, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	defer f.Close()

	if err != nil {
		fmt.Print(err)
	}
}

func cleanupFile(filePath string, t *testing.T) {
	if err := os.Remove(filePath); err != nil {
		t.Error("Cant remove file ", filePath, err)
		t.Fail()
	}
}
