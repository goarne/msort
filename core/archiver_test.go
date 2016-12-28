package core

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"
)

var (
	testfileRootDirectory        = "./testfolder/"
	testFileName          string = "testfilename"
	ArchiverFileChannel   chan *ArchiveFile
)

func TestSuite(t *testing.T) {
	setCmmdPrmForArchiverUnderTest(testFileName)

	createTestFile(testfileRootDirectory, testFileName)
	defer os.RemoveAll(testfileRootDirectory)

	setUpArchiver(2)
	t.Run("Should find file and archive file (integrationtest).", shouldFindAndArchiveFileAsync)

	setUpArchiver(1)
	t.Run("Should find file.", shouldFindFile)

	setUpArchiver(1)
	t.Run("Should not find file", shouldNotFindFile)

	setUpArchiver(1)
	t.Run("Shall archive file", shouldArchiveFile)
}

func shouldFindAndArchiveFileAsync(t *testing.T) {

	defer func() {
		if r := recover(); r == nil {
			t.Error("Channel was not closed.")
		}
	}()

	go FindFiles(ArchiverFileChannel)
	go ArchiveFiles(ArchiverFileChannel)
	WaitAsync()

	if fn, _ := archivedFileExists(testFileName); fn == nil {
		t.Fail()
	}

	//The statement will cause a panic and make the defered function fail
	ArchiverFileChannel <- &ArchiveFile{}
}

func shouldArchiveFile(t *testing.T) {
	af := &ArchiveFile{}
	af.FileInfo, _ = os.Stat(testfileRootDirectory + testFileName)
	af.SourcePath = CmdPrm.Source + testFileName

	go func() {
		ArchiverFileChannel <- af
		defer close(ArchiverFileChannel)
	}()

	go ArchiveFiles(ArchiverFileChannel)

	if fn, _ := archivedFileExists(testFileName); fn == nil {
		t.Fail()
	}
}

func shouldFindFile(t *testing.T) {

	go FindFiles(ArchiverFileChannel)

	if testFileFound := <-ArchiverFileChannel; testFileFound == nil {
		t.Fail()
	}
}

func shouldNotFindFile(t *testing.T) {
	CmdPrm.FilePattern = "patternNotMatchingTestfile"

	go FindFiles(ArchiverFileChannel)

	if testFileFound := <-ArchiverFileChannel; testFileFound != nil {
		t.Fail()
	}
}

//Below are helper functions for setting up unittest and verifying results.
func setCmmdPrmForArchiverUnderTest(fn string) {
	CmdPrm.Source = testfileRootDirectory
	CmdPrm.Target = testfileRootDirectory
	CmdPrm.FilePattern = fn
	CmdPrm.TargetPattern = "YYYY/MM/DD"
	CmdPrm.ShallArchive = true
	CmdPrm.Overwrite = false
}

func archivedFileExists(fn string) (os.FileInfo, error) {

	year := strconv.Itoa(time.Now().Year())
	month := strconv.Itoa(int(time.Now().Month()))
	day := strconv.Itoa(time.Now().Day())

	return os.Stat(testfileRootDirectory + year + "/" + month + "/" + day + "/" + fn)
}

func setUpArchiver(numAsync int) {
	ArchiverFileChannel = make(chan *ArchiveFile, 1)
	StartAsync(numAsync)
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
