package internal

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/sys/unix"
)

func initMountNamespace() (string, error) {
	err := unix.Unshare(unix.CLONE_NEWNS)
	if err != nil {
		return "", fmt.Errorf("Error creating new mount namespace: %v", err)
	}
	tmpPath, err := os.MkdirTemp("", "")
	if err != nil {
		return "", fmt.Errorf("Error creating scratch mountpoint: %v", err)
	}
	err = unix.Mount("swap", tmpPath, "tmpfs", 0, "")
	if err != nil {
		return tmpPath, fmt.Errorf("Error mounting namespace scratch: %v", err)
	}
	return tmpPath, nil
}

func cleanMountNamespace(scratchPath string, overlays []FileOverlay) {
	for _, overlay := range overlays {
		err := unix.Unmount(overlay.Path, unix.MNT_FORCE)
		if err != nil {
			log.Printf("Error unmounting %s: %v", overlay.Path, err)
		}
	}
	err := unix.Unmount(scratchPath, 0)
	if err != nil {
		log.Printf("Error unmounting scratch: %v", err)
		return
	}
	err = os.Remove(scratchPath)
	if err != nil {
		log.Printf("Error removing scratch mountpoint: %v", err)
		return
	}
}

func setupMounts(scratchPath string, mounts []FileOverlay) {
	for _, overlay := range mounts {
		f, err := os.CreateTemp(scratchPath, "")
		if err != nil {
			log.Printf("Error rendering scratch source for %s: %v", overlay.Path, err)
			continue
		}
		n, err := f.WriteString(overlay.Content)
		if err != nil {
			log.Printf("Error creating overlay file: wrote %d bytes, then error: %v", n, err)
		}
		srcPath := f.Name()
		f.Close()
		err = unix.Mount(srcPath, overlay.Path, "", unix.MS_BIND|unix.MS_PRIVATE|unix.MS_REC, "")
		if err != nil {
			log.Printf("Error bind mounting %s: %v", overlay.Path, err)
			continue
		}
	}
}
