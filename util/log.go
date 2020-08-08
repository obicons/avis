package util

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var file *os.File
var mut sync.Mutex

func NewTempLogger() (*os.File, error) {
	mut.Lock()
	defer mut.Unlock()
	if file == nil {
		filename := fmt.Sprintf("/tmp/rmck-%d", time.Now().Unix())
		var err error
		file, err = os.Create(filename)
		if err != nil {
			return nil, err
		}
	}
	return file, nil
}
