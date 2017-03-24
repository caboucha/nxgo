package nx

import (
        "bytes"
        "fmt"
)

func (c *Client) AddVlan(vlanId string, vni string) error {

    var segment string

    if vni !=  "" {
        segment = fmt.Sprintf(vxlanSegment, vni)
    } else {
        segment = ""
    }
      
    result := fmt.Sprintf(vlanEntity, vlanId, segment)

    jsonVlan := TopBegin+result+TopEnd
    c.debugf("vlan add: Body=%s", jsonVlan)

    body, errPost := c.post(ConfigRootURI, contentTypeJSON,
                            bytes.NewBufferString(jsonVlan))
    if errPost != nil {
        return errPost
    }

    return parseJSONError(body)
}


func (c *Client) GetVlan(id string) ([]map[string]interface{}, error) {
    var uri string

    if id == "" {
        uri = AllVlanURI
    } else {
        uri = fmt.Sprintf(VlanURI, id)
    }

    body, errGet := c.get(uri)
    if errGet != nil {
            return nil, errGet
    }

    return jsonImdataAttributes(c, body, "l2BD", c.getFuncName(1))

}

func (c *Client) DeleteVlan(id string) (error) {
    var uri string

    uri = fmt.Sprintf(VlanURI, id)

    body, errDel := c.delete(uri)
    if errDel != nil {
            return errDel
    }

    return parseJSONError(body)
}

