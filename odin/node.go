package odin

import (
	"encoding/base64"
	"encoding/json"
)

// Node represents a single entry in the deserialized Odin data tree.
// It preserves all metadata needed for perfect round-trip serialization.
type Node struct {
	// BinType is the exact BinaryEntryType byte that produced this node.
	BinType BinaryEntryType `json:"binType"`

	// Name is the field/entry name (empty for unnamed entries).
	Name string `json:"name,omitempty"`

	// TypeInfo holds the type metadata for node entries.
	TypeInfo *TypeInfo `json:"typeInfo,omitempty"`

	// NodeID is the reference ID for reference nodes (-1 for struct nodes).
	NodeID int32 `json:"nodeId,omitempty"`

	// ArrayLength for StartOfArray entries.
	ArrayLength int64 `json:"arrayLength,omitempty"`

	// Value holds the primitive value for leaf entries.
	// Types: int8, uint8, int16, uint16, int32, uint32, int64, uint64,
	//        float32, float64, string, bool, nil, []byte (for primitive arrays / guid / decimal)
	Value interface{} `json:"value,omitempty"`

	// PrimArrayInfo holds metadata for PrimitiveArray entries.
	PrimArrayInfo *PrimArrayInfo `json:"primArrayInfo,omitempty"`

	// Children holds child nodes for node/array entries.
	Children []*Node `json:"children,omitempty"`
}

// UnmarshalJSON implements custom unmarshaling to preserve types in the Value interface{}.
func (n *Node) UnmarshalJSON(data []byte) error {
	type Alias Node
	var a Alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*n = Node(a)

	if n.Value == nil {
		return nil
	}

	// Based on BinType, determine if Value needs special conversion from Go's default json.Unmarshal (string or float64 or map).
	switch n.BinType {
	case PrimitiveArray, NamedExternalReferenceByGuid, UnnamedExternalReferenceByGuid, NamedDecimal, UnnamedDecimal, NamedGuid, UnnamedGuid:
		if vStr, ok := n.Value.(string); ok {
			var bw ByteSliceWrapper
			// ByteSliceWrapper string is unmarshaled manually
			if err := json.Unmarshal([]byte(`"`+vStr+`"`), &bw); err == nil {
				n.Value = bw
			}
		} else if vMap, ok := n.Value.(map[string]interface{}); ok {
			// Possibly a nested Node
			if _, hasBinType := vMap["binType"]; hasBinType {
				b, err := json.Marshal(vMap)
				if err == nil {
					var subNode Node
					if err := json.Unmarshal(b, &subNode); err == nil {
						n.Value = &subNode
					}
				}
			}
		}
	case NamedSByte, UnnamedSByte, NamedByte, UnnamedByte:
		if f, ok := n.Value.(float64); ok {
			if n.BinType == NamedSByte || n.BinType == UnnamedSByte {
				n.Value = int8(f)
			} else {
				n.Value = byte(f)
			}
		}
	}

	return nil
}

// TypeInfo stores serialized type metadata.
type TypeInfo struct {
	// IsTypeID is true if this type was encoded as a TypeID reference.
	IsTypeID bool `json:"isTypeId,omitempty"`
	// TypeName is the full .NET type name (only set for TypeName entries).
	TypeName string `json:"typeName,omitempty"`
	// TypeIDValue is the numeric type ID.
	TypeIDValue int32 `json:"typeIdValue,omitempty"`
}

// PrimArrayInfo stores metadata for PrimitiveArray entries.
type PrimArrayInfo struct {
	ElementCount   int32 `json:"elementCount"`
	BytesPerElement int32 `json:"bytesPerElement"`
}

// ByteSliceWrapper wraps []byte for JSON serialization as base64.
type ByteSliceWrapper struct {
	Data []byte
}

func (b ByteSliceWrapper) MarshalJSON() ([]byte, error) {
	encoded := base64.StdEncoding.EncodeToString(b.Data)
	return []byte(`"` + encoded + `"`), nil
}

func (b *ByteSliceWrapper) UnmarshalJSON(data []byte) error {
	// Strip quotes
	s := string(data)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	b.Data = decoded
	return nil
}
