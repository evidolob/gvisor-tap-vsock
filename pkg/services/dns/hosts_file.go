package dns

import (
	"net"
	"os"
	"sync"
	"time"

	"github.com/areYouLazy/libhosty"
)

type HostsFile interface {
	LookupByHostname(name string) (net.IP, error)
	BachOperation(started bool)
}

type hosts struct {
	lastModified              time.Time
	hostsReadLock             sync.RWMutex
	hostsFilePath             string
	hostsFile                 *libhosty.HostsFile
	noNeedToCheckModification bool
}

// NewHostsFile Creates new HostsFile instance
// Pass ""(empty string) if you want to use default hosts file
func NewHostsFile(hostsPath string) (HostsFile, error) {
	hostsFile, err := readHostsFile(hostsPath)
	if err != nil {
		return nil, err
	}

	return &hosts{
		hostsFile:     hostsFile,
		hostsFilePath: hostsFile.Config.FilePath,
		lastModified:  time.Now(),
	}, nil
}

func (h *hosts) LookupByHostname(name string) (net.IP, error) {
	if !h.noNeedToCheckModification {
		err := h.checkAndReadHosts()
		if err != nil {
			return nil, err
		}
	}

	_, ip, err := h.hostsFile.LookupByHostname(name)
	return ip, err
}

func (h *hosts) checkAndReadHosts() error {
	h.hostsReadLock.RLock()
	defer h.hostsReadLock.RUnlock()
	hostsStat, err := os.Stat(h.hostsFilePath)
	if err != nil {
		return err
	}

	if h.lastModified.Before(hostsStat.ModTime()) {
		newHosts, err := readHostsFile(h.hostsFilePath)
		if err != nil {
			return err
		}
		h.hostsFile = newHosts
		h.lastModified = hostsStat.ModTime()
	}
	return nil
}

func (h *hosts) BachOperation(started bool) {
	h.hostsReadLock.RLock()
	defer h.hostsReadLock.RUnlock()
	h.noNeedToCheckModification = started
}

func readHostsFile(hostsFilePath string) (*libhosty.HostsFile, error) {
	config, err := libhosty.NewHostsFileConfig(hostsFilePath)
	if err != nil {
		return nil, err
	}
	return libhosty.InitWithConfig(config)
}
