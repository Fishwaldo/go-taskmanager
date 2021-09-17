package utils

import (
    "testing"

    "github.com/stretchr/testify/require"
)

func TestCopyMap(t *testing.T) {
    m1 := map[string]interface{}{
        "a": "bbb",
        "b": map[string]interface{}{
            "c": 123,
        },
        "c": []interface{} {
            "d", "e", map[string]interface{} {
                "f": "g",
            },
        },
    }

    m2 := CopyableMap(m1).DeepCopy()

    m1["a"] = "zzz"
    delete(m1, "b")
    m1["c"].([]interface{})[1] = "x"
    m1["c"].([]interface{})[2].(map[string]interface{})["f"] = "h"

    require.Equal(t, map[string]interface{}{
        "a": "zzz", 
        "c": []interface{} {
            "d", "x", map[string]interface{} {
                "f": "h",
            },
        },
    }, m1)
    require.Equal(t, map[string]interface{}{
        "a": "bbb",
        "b": map[string]interface{}{
            "c": 123,
        },
        "c": []interface{} {
            "d", "e", map[string]interface{} {
                "f": "g",
            },
        },
    }, m2)
}