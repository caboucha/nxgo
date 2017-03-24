package nx

import (
        "bytes"
        "fmt"
        "strings"
)

// SplitInterfaceName separates the interface type and id from interface name
// Ex: ethernet:1/12 results in two fields: ethernet 1/12
func (c *Client) SplitInterfaceName(ifName string) (string, string, error) {

    var id string

    intf_split := strings.Split(ifName, ":")

    size := len(intf_split)
    if size != 2  && size != 1 {
        getErr := fmt.Errorf(`ERROR: Unexpected interface value %s.
                             Example Values: ethernet:1/3 or port-channel:5
                             or ethernet or port-channel`,
                             ifName)
        return "", "", getErr
    }

    if size == 1 {
        id = ""
    } else {
        id = intf_split[1]
    }

    return intf_split[0], id, nil

}

// formatTrunkBody - formats the json body of add/replace operations
// for adding trunk/native vlans to interface
func (c *Client) formatTrunkBody(iftype string, id string, 
    allowed string, native string) (string, error) {
    var vlancfg string
    var tag string
    var pfx string

    if allowed == "None" {
        allowed = ""
    }
    if native == "" {
        vlancfg = fmt.Sprintf(TrunkVlans, allowed)
    } else {
        if native == "None" {
            native = "1"
        }
        
        s := []string{fmt.Sprintf(NativeVlan, native),
                      fmt.Sprintf(TrunkVlans, allowed)}
        vlancfg = strings.Join(s, ", ")
    }

    switch iftype {
    case "ethernet":
        fallthrough
    case "enet":
        tag = EnetTag
        pfx = EnetPfx
    case "port-channel":
        fallthrough
    case "po":
        tag = PcTag
        pfx = PcPfx
    default:
        return "", fmt.Errorf("Unexpected interface type: %s", iftype)
    }

    result := fmt.Sprintf(IfEntity, tag, pfx, id,
                          TrunkMode, vlancfg)
    return TopBegin+result+TopEnd, nil
}

// AddTrunkVlan - Adds trunk/native Vlan to interface
func (c *Client) AddTrunkVlan(ifName string, 
                              allowed string, 
                              native string) error {

        ifType, ifId, err := c.SplitInterfaceName(ifName)
        if err != nil {
            return err
        }

        jsonTrunk, err := c.formatTrunkBody(ifType, ifId, allowed, native)
        if err != nil {
            return err
        }

        c.debugf("Ethernet trunk vlan add: Body=%s", jsonTrunk)

        body, errPost := c.post(ConfigRootURI, contentTypeJSON, 
                                bytes.NewBufferString(jsonTrunk))
        if errPost != nil {
                return errPost
        }

        return parseJSONError(body)
}

func (c *Client) GetInterface(ifName string) ([]map[string]interface{}, error) {

    var uri, urifmt, key string

    pfx, id, err := c.SplitInterfaceName(ifName)
    if err != nil {
        return nil, err
    }

    switch pfx {
    case "ethernet":
        key = "l1PhysIf"
        urifmt = InterfaceEnetURI
    case "port-channel":
        key = "pcAggrIf"
        urifmt = InterfacePcURI

    default:
        errGet := fmt.Errorf(`ERROR: Unexpected Interface name %s for get.
                             Example Values: ethernet:1/3 or port-channel:5
                             or ethernet or port-channel`,
                             ifName)
        return nil, errGet
    }
    if id == "" {
        uri = fmt.Sprintf(InterfaceAll, key)
    } else {
        uri = fmt.Sprintf(urifmt, id)
    }

    body, errGet := c.get(uri)
    if errGet != nil {
            return nil, errGet
    }
    
    return jsonImdataAttributes(c, body, key, c.getFuncName(1))
}

