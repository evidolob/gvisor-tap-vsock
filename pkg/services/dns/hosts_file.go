package dns

import (
	"net"
	"sync"

	"github.com/areYouLazy/libhosty"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

type HostsFile interface {
	LookupByHostname(name string) (net.IP, error)
}

type hosts struct {
	hostsReadLock sync.RWMutex
	hostsFilePath string
	hostsFile     *libhosty.HostsFile
}

// NewHostsFile Creates new HostsFile instance
// Pass ""(empty string) if you want to use default hosts file
func NewHostsFile(hostsPath string) (HostsFile, error) {
	hostsFile, err := readHostsFile(hostsPath)
	if err != nil {
		return nil, err
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	h := &hosts{
		hostsFile:     hostsFile,
		hostsFilePath: hostsFile.Config.FilePath,
	}
	go func() {
		h.startWatch(watcher)
	}()
	return h, nil
}

func (h *hosts) startWatch(w *fsnotify.Watcher) {
	err := w.Add(h.hostsFilePath)
	if err != nil {
		log.Errorf("Hosts file adding watcher error:%s", err)
		return
	}
	for {
		select {
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Errorf("Hosts file watcher error:%s", err)
		case event, ok := <-w.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				err := h.updateHostsFile()
				if err != nil {
					log.Errorf("Hosts file read error:%s", err)
					return
				}
			}
		}
	}
}

func (h *hosts) LookupByHostname(name string) (net.IP, error) {
	_, ip, err := h.hostsFile.LookupByHostname(name)
	return ip, err
}

func (h *hosts) updateHostsFile() error {
	h.hostsReadLock.RLock()
	defer h.hostsReadLock.RUnlock()
	newHosts, err := readHostsFile(h.hostsFilePath)
	if err != nil {
		return err
	}
	h.hostsFile = newHosts
	return nil
}

func readHostsFile(hostsFilePath string) (*libhosty.HostsFile, error) {
	config, err := libhosty.NewHostsFileConfig(hostsFilePath)
	if err != nil {
		return nil, err
	}
	return libhosty.InitWithConfig(config)
}
