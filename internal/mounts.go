package internal

import (
    "golang.org/x/sys/unix"
    "os"
    "fmt"
    "log"
)

func initMountNamespace() (string, error) {
    unix.Unshare(unix.CLONE_NEWNS)
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
    return
}

func setupMounts(scratchPath string, mounts []FileOverlay) {
    for _, overlay := range mounts {
        f, err := os.CreateTemp(scratchPath, "")
        if err != nil {
            log.Printf("Error rendering scratch source for %s: %v", overlay.Path, err)
            continue
        }
        f.WriteString(overlay.Content)
        srcPath := f.Name()
        f.Close()
        err = unix.Mount(srcPath, overlay.Path, "", unix.MS_BIND | unix.MS_PRIVATE | unix.MS_REC, "")
        if err != nil {
            log.Printf("Error bind mounting %s: %v", overlay.Path, err)
            continue
        }
    }
}
