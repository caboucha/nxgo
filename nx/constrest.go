package nx

const (
    // Single URI Definition for Add, Replace, interface Delete attributes
    ConfigRootURI = "/api/mo.json"

    // Body definitions Start
    TopBegin = `{"topSystem": { "children": [ `
    TopEnd = `]}}`

    // URI Definition for Get
    // Where %s is 'l1PhysIf' for Enet and 'pcAggrIf' for Port Channel
    InterfaceAll = "/api/mo/sys/intf.json?query-target=subtree&target-subtree-class=%s"

    // Where %s is interface id. Ex: 1/12 for ethernet or 55 for port-channel
    InterfaceEnetURI = "/api/mo/sys/intf/phys-[eth%s].json"
    InterfacePcURI = "/api/mo/sys/intf/aggr-[po%s].json"

    // switchport mode <which-mode> Where id is interface-id ex: po5 eth1/3
    // modes are trunk, access, or edge
    SwitchPortMode = `{ "stpEntity": { "children": [ {
                     "stpInst": { "children": [ { "stpIf": {
                     "attributes": { "id": "%s",
                     "mode": "%s" } } } ] } } ] } }`
    TrunkMode = "trunk"
    AccessMode = "access"
    EdgeMode = "edge"

    // switchport trunk native vlan 129
    // switchport trunk allowed vlan 129,136
    // 1st %s is pcAggrIf for Port channel and l1PhysIf for ethernet
    // 2nd&3rd %s is interface-id ex: po5 eth1/3
    // 4th %s is TrunkMode (above)
    // 5th %s is trunkVlans and nativeVlan config (below)
    IfEntity = `{ "interfaceEntity": { "children": [ { "%s": { "attributes": { "id": "%s%s", "mode": "%s", %s } } } ] } }`
    PcTag = "pcAggrIf"
    EnetTag = "l1PhysIf"
    PcPfx = "po"
    EnetPfx = "eth"
    NativeVlan = `"nativeVlan": "vlan-%s"`
    NoNativeVlan = `"nativeVlan": ""`
    TrunkVlans = `"trunkVlans": "%s"`

    //  Creates vpc on the Port Channel itself
    //    int port-channel 19
    //        vpc 19
    vpcEntity = `{ "vpcEntity": { "children": [ { "vpcInst": {
                 "children": [ { "vpcDom": {
                 "children": [ { "vpcIf": {
                 "attributes": { "id": "%s" },
                 "children": [ { "vpcRsVpcConf": {
                 "attributes": { "tDn": "sys/intf/aggr-[po%s]"
                 } } } ] } } ] } } ] } } ] } }`

    // URI Definition for Get, Delete
    VlanURI = `/api/mo/sys/bd/bd-[vlan-%s].json`
    AllVlanURI = `/api/mo/sys/bd/.json?query-target=subtree&target-subtree-class=l2BD`

    // VLAN Body for add operation
    vlanEntity = `{ "bdEntity": { "children": [ {"l2BD": {"attributes": {"fabEncap": "vlan-%s", "pcTag": "1", "adminSt": "active"%s } } } ] } }`
    vxlanSegment = `, "accEncap": "vxlan-%s"`

)

