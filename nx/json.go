package nx

import (
	"fmt"
        "encoding/json"
)

func mapGet(i interface{}, member string) (interface{}, error) {
	m, isMap := i.(map[string]interface{})
	if !isMap {
		return nil, fmt.Errorf("json mapGet: not a map")
	}
	mem, found := m[member]
	if !found {
		return nil, fmt.Errorf("json mapGet: member [%s] not found", member)
	}
	return mem, nil
}

func sliceGet(i interface{}, member int) (interface{}, error) {
	list, isList := i.([]interface{})
	if !isList {
		return nil, fmt.Errorf("json sliceGet: not a slice")
	}
	if member < 0 || member >= len(list) {
		return nil, fmt.Errorf("json sliceGet: member=%d out-of-bounds", member)
	}
	return list[member], nil
}

func mapSimple(i interface{}, member string) interface{} {
	m, _ := mapGet(i, member)
	return m
}

func mapString(i interface{}, member string) string {
	m := mapSimple(i, member)
	s, isStr := m.(string)
	if isStr {
		return s
	}
	return ""
}
func parseJSONError(body []byte) error {

        var reply interface{}
        errJSON := json.Unmarshal(body, &reply)
        if errJSON != nil {
                return errJSON
        }

        imdata, imdataError := mapGet(reply, "imdata")
        if imdataError != nil {
                return fmt.Errorf("imdata error: %s", string(body))
        }

        list, isList := imdata.([]interface{})
        if !isList {
                return fmt.Errorf("imdata does not hold a list: %s", string(body))
        }

        if len(list) == 0 {
                return nil // ok
        }

        first := list[0]

        e, errErr := mapGet(first, "error")
        if errErr != nil {
                return nil // ok
        }

        attr := mapSimple(e, "attributes")
        code := mapString(attr, "code")
        text := mapString(attr, "text")

        return fmt.Errorf("error: code=%s text=%s", code, text)
}
