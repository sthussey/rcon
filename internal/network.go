package internal

import (
    "fmt"
    "log"
    "math/rand"
    "net"
    "time"
    ulid "github.com/oklog/ulid/v2"
    "github.com/vishvananda/netlink"
    "github.com/vishvananda/netns"
    "inet.af/netaddr"
)

type ProcNetwork struct {
    cidr            netaddr.IPPrefix
    gateway         netaddr.IP
    baseNlHandle    *netlink.Handle
    procNlHandle    *netlink.Handle
    baseNsHandle    netns.NsHandle
    procNsHandle    netns.NsHandle
    rootLink        netlink.Link
    peerLink        netlink.Link
    rootAddr        netlink.Addr
    peerAddr        netlink.Addr
    teardown        *NetworkTeardown
}

type NetworkTeardown struct {
    teardownFuncs []func()
}

func CreateNetworkTeardown() *NetworkTeardown {
    t := NetworkTeardown{teardownFuncs: make([]func(), 0)}
    return &t
}

func (t *NetworkTeardown) teardown() {
    for i := len(t.teardownFuncs)-1; i >= 0; i-- {
        t.teardownFuncs[i]()
    }
}

func (t *NetworkTeardown) add(f func()) *NetworkTeardown {
    t.teardownFuncs = append(t.teardownFuncs, f)
    return t
}

func generateNetNSName() (string, error) {
    entropySrc := rand.New(rand.NewSource(time.Now().UnixNano()))
    name, err := ulid.New(ulid.Timestamp(time.Now()), entropySrc)
    if err != nil {
        return "", fmt.Errorf("Error generating name: %v", err)
    }
    log.Printf("Replacing name %s with 'footest'", name.String())
    return "footest", nil
}

func createLinkedNetNS(name string) (*ProcNetwork, error) {
    procNetwork := ProcNetwork{teardown: CreateNetworkTeardown()}
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

    procNetwork.teardown.add(func(){procNetwork.procNsHandle.Close()
                                    netns.DeleteNamed(name)})

    if procNetwork.procNlHandle, err = netlink.NewHandleAt(procNetwork.procNsHandle); err != nil {
        procNetwork.teardown.teardown()
        return nil, fmt.Errorf("Error getting proc netlink handle: %v", err)
    }

    procNetwork.teardown.add(func(){procNetwork.procNlHandle.Delete()})

    var rootLink *netlink.Veth
    rootLinkMAC, _ := net.ParseMAC("EE:EE:EE:EE:EE:EE")
    peerLinkMAC, _ := net.ParseMAC("EE:EE:EE:EE:EE:EF")

    rootLinkAttrs := netlink.NewLinkAttrs()
    rootLinkAttrs.Name = fmt.Sprintf("%s-link", name)
    rootLinkAttrs.HardwareAddr = rootLinkMAC
    rootLinkAttrs.Flags = net.FlagUp | net.FlagBroadcast | net.FlagMulticast
    rootLinkAttrs.MTU = 1500
    rootLinkAttrs.TxQLen = 1000
    rootLinkAttrs.NumTxQueues = 8
    rootLinkAttrs.NumRxQueues = 8
    rootLinkAttrs.Namespace = netlink.NsFd(procNetwork.baseNsHandle)
    rootLinkAttrs.NetNsID = int(procNetwork.procNsHandle)

    rootLink = &netlink.Veth{
        LinkAttrs: rootLinkAttrs,
        PeerName: fmt.Sprintf("%s-peer", name),
        PeerHardwareAddr: peerLinkMAC,
        PeerNamespace: netlink.NsFd(procNetwork.procNsHandle)}
    log.Printf("Created rootlink %v", rootLink)

    if err := procNetwork.baseNlHandle.LinkAdd(rootLink); err != nil {
        procNetwork.teardown.teardown()
        return nil, fmt.Errorf("Error creating veth link: %v", err)
    } else {
        procNetwork.rootLink = rootLink
        procNetwork.peerLink, err = procNetwork.procNlHandle.LinkByName(rootLink.PeerName)
        if err != nil {
            procNetwork.teardown.teardown()
            return nil, fmt.Errorf("Error getting peer link: %v", err)
        }
        procNetwork.teardown.add(func(){procNetwork.baseNlHandle.LinkDel(procNetwork.rootLink)
                                        procNetwork.procNlHandle.LinkDel(procNetwork.peerLink)})
    }

    err = netlink.LinkSetUp(procNetwork.peerLink)

    return &procNetwork, nil
}


func configureNetNS(config NetworkConfig, nw *ProcNetwork) error {
    cidr, err := netaddr.ParseIPPrefix(config.NamespaceCidr)
    if err != nil {
        return fmt.Errorf("Error parsing CIDR %s: %v", config.NamespaceCidr, err)
    }
    if cidr.Bits > 30 {
        return fmt.Errorf("Error: namespace CIDR %s is invalid, does not contain minimum IPs (4).", config.NamespaceCidr)
    }
    iprange := cidr.Range()
    if (!iprange.Valid()) {
        return fmt.Errorf("Error: could not derive IP range from CIDR")
    }
    gw_addr := netlink.Addr{IPNet: &net.IPNet{IP: iprange.From.Next().IPAddr().IP, Mask: net.CIDRMask(int(cidr.Bits), 32)}}
    ns_addr := netlink.Addr{IPNet: &net.IPNet{IP: iprange.From.Next().Next().IPAddr().IP, Mask: net.CIDRMask(int(cidr.Bits), 32)}}
    err = nw.baseNlHandle.AddrAdd(nw.rootLink, &gw_addr)
    if err != nil {
        return fmt.Errorf("Error: could not create address on root link: %v", err)
    }
    nw.rootAddr = gw_addr
    err = nw.procNlHandle.AddrAdd(nw.peerLink, &ns_addr)
    if err != nil {
        return fmt.Errorf("Error: could not create address on peer link: %v", err)
    }
    nw.peerAddr = ns_addr
    route := netlink.Route{
                Scope: netlink.SCOPE_UNIVERSE,
                LinkIndex: nw.peerLink.Attrs().Index,
                Dst: nil,
                Gw: gw_addr.IP}
    netns.Set(nw.procNsHandle)
    err = nw.procNlHandle.RouteAdd(&route)
    if err != nil {
        return fmt.Errorf("Error adding default route in netns: %v", err)
    }
    return nil
}

func setupNetwork(config NetworkConfig) (*ProcNetwork, error) {
    if config.NamespaceCidr == "" {
        return nil, nil
    }
    pn, err := createLinkedNetNS("")
    if err != nil {
        return nil, fmt.Errorf("Error creating netns: %v", err)
    }
    log.Printf("Process Network Created:\n %v", *pn)
    err = configureNetNS(config, pn)
    if err != nil {
        pn.teardown.teardown()
        return nil, fmt.Errorf("Error configuring netns: %v", err)
    }
    pn.teardown.add(func(){netns.Set(pn.baseNsHandle)})
    return pn, nil
}

