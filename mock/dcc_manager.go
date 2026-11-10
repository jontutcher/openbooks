package mock

import (
	"github.com/evan-buss/openbooks/dcc"
	"net"
	"os"
	"path/filepath"
)

type DccManager struct{}

func NewDccManager() *DccManager {
	return &DccManager{}
}

// ServeFile starts a one-shot DCC server for path and returns the DCC SEND
// string a client needs in order to download it.
func (d *DccManager) ServeFile(path string) (string, error) {
	file, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	server := NewDccServer(path)
	selectedPortChan := make(chan int)
	go server.Start(selectedPortChan)

	return dcc.New(filepath.Base(path), net.ParseIP("127.0.0.1").To4(), <-selectedPortChan, file.Size()).String(), nil
}
