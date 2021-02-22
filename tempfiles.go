package daylight

import (
	"io/ioutil"
	"os"
	"sync"
)

// TempFiles is a collection of temporary files that can be removed on some
// sufficiently slow schedule.
//
// In daylight this is used to manage .ics files which need to be read by a
// calendar application before they can be deleted. Temp file cleanup can race
// with the calendar application, but since cleanups are infrequent, calendar
// event creation is even less frequent, and calendar apps shouldn't take a
// long time to read an ICS file, this risk condition is tolerable.
type TempFiles struct {
	sync.Mutex
	FileNameFormat string
	files          []*os.File
}

func (fs *TempFiles) append(file *os.File) {
	fs.Lock()
	defer fs.Unlock()
	fs.files = append(fs.files, file)
}

// CleanUp removes all files in fs.
func (fs *TempFiles) CleanUp() {
	fs.Lock()
	defer fs.Unlock()
	for _, f := range fs.files {
		os.Remove(f.Name())
	}
	fs.files = nil
}

// New returns a new TempFile with fs's name format. That file belongs to fs,
// and will be removed when fs.CleanUp() is called.
func (fs *TempFiles) New() (*os.File, error) {
	tmpfile, err := ioutil.TempFile("", fs.FileNameFormat)
	fs.append(tmpfile)
	return tmpfile, err
}
