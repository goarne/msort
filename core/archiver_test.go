package core

import (
	"test"
)

var appConfig AppConfig

func shouldFindPropertiesFile(t *testing.T) {
	appConfig = Ap
	fileList := make(chan *core.ArchiveFile)

	core.FindFiles(fileList)

	f <- fileList

}
