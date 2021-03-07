package internal

import (
    "fmt"
    "math/rand"
    "time"
    ulid "github.com/oklog/ulid/v2"
    "github.com/vishvananda/netlink"
    "github.com/vishvananda/netns"
)

type ProcNetwork struct {
    baseNlHandle    *netlink.Handle
    procNlHandle    *netlink.Handle
    baseNsHandle    netns.NsHandle
    procNsHandle    netns.NsHandle
    rootLink    *netlink.Link
    peerLink    *netlink.Link
    teardown    *NetworkTeardown
}

type NetworkTeardown struct {
    teardownFuncs []func()
}

func CreateNetworkTeardown() *NetworkTeardown {
    t := NetworkTeardown{teardownFunc: make([]func(), 0)
    return &t
}

func (t *NetworkTeardown) teardown() {
    for _, f := range t {
        f()
    }
}

func (t *NetworkTeardown) add(f func()) *NetworkTeardown {
    t.teardownFunc = append(t.teardownFunc, f)
    return t
}

func generateNetNSName() (string, error) {
    entropySrc := rand.New(rand.NewSource(time.Now().UnixNano()))
    name, err := ulid.New(ulid.Timestamp(time.Now()), entropySrc)
    if err != nil {
        return "", fmt.Errorf("Error generating name: %v", err)
    }
    return name.String(), nil
}

func createLinkedNetNS(name string, linkbase string) (*ProcNetwork, error) {
    procNetwork := ProcNetwork{teardown: CreateNetworkTeardown}
    var err error

    if procNetwork.baseNsHandle, err = netns.Get(); err != nil {
        return nil, fmt.Errorf("Error getting base netns: %v", err)
    }

    if procNetwork.baseNlHandle, err = netlink.NewHandle(); err != nil {
        return nil, fmt.Errorf("Error getting base netlink handle: %v", err)
    }

    procNetwork.teardown.add(func(){procNetwork.baseNlHandle.Delete()})

    if name == "" {
        name, err = generateNetNSName()
        if err != nil {
            procNetwork.teardown.teardown()
            return nil, fmt.Errorf("Error creating netns: %v", err)
        }
    }

    procNetwork.procNsHandle, err = netns.NewNamed(name)

    if err != nil {
        return nil, fmt.Errorf("Error creating netns: %v", err)
    }

    procNetwork.teardown.add(func(){procNetwork.procNsHandle.Close()})

    if procNetwork.procNlHandle, err = netlink.NewHandleAt(procNetwork.procNsHandle); err != nil {
        procNetwork.teardown.teardown()
        return nil, fmt.Errorf("Error getting proc netlink handle: %v", err)
    }

    procNetwork.teardown.add(func(){procNetwork.procNlHandle.Delete())

    var rootLink *netlink.Veth

    if localNS, err := netns.Get(); err == nil {
        rootLink = &netlink.Veth{netlink.LinkAttrs{Name: fmt.Sprintf("%s-link", name), Namespace: netlink.NsFd(localNS)}, fmt.Sprintf("%s-peer", name), nil, procNetwork.nsHandle}
    } else {
        procNetwork.teardown.teardown()
        return nil, fmt.Errorf("Error getting current network namespace: %v", err)
    }

    if err := netlink.LinkAdd(rootLink); err != nil {
        procNetwork.teardown.teardown()
        return nil, fmt.Errorf("Error creating veth link: %v", err)
    }


    return &procNetwork, nil
}


func configureNetNS(config ProcWrapNetworkConfig, nw *ProcNetwork) error {

}


