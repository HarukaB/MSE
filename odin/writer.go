package odin

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"
)

// BinaryDataWriter writes Odin Serializer binary format.
type BinaryDataWriter struct {
	buf []byte
}

// NewBinaryDataWriter creates a new writer.
func NewBinaryDataWriter() *BinaryDataWriter {
	return &BinaryDataWriter{}
}

// Bytes returns the written data.
func (w *BinaryDataWriter) Bytes() []byte {
	return w.buf
}

// ---- low-level write helpers ----

func (w *BinaryDataWriter) writeByte(b byte) {
	w.buf = append(w.buf, b)
}

func (w *BinaryDataWriter) writeBytes(data []byte) {
	w.buf = append(w.buf, data...)
}

func (w *BinaryDataWriter) writeInt32(v int32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], uint32(v))
	w.buf = append(w.buf, b[:]...)
}

func (w *BinaryDataWriter) writeUint32(v uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	w.buf = append(w.buf, b[:]...)
}

func (w *BinaryDataWriter) writeInt64(v int64) {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], uint64(v))
	w.buf = append(w.buf, b[:]...)
}

func (w *BinaryDataWriter) writeUint64(v uint64) {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], v)
	w.buf = append(w.buf, b[:]...)
}

func (w *BinaryDataWriter) writeFloat32(v float32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], float32ToBits(v))
	w.buf = append(w.buf, b[:]...)
}

func (w *BinaryDataWriter) writeFloat64(v float64) {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], float64ToBits(v))
	w.buf = append(w.buf, b[:]...)
}

func (w *BinaryDataWriter) writeInt16(v int16) {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], uint16(v))
	w.buf = append(w.buf, b[:]...)
}

func (w *BinaryDataWriter) writeUint16(v uint16) {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], v)
	w.buf = append(w.buf, b[:]...)
}

// writeString writes an Odin-encoded string.
// Force 16-bit encoding for exact bit-for-bit match with game's output.
func (w *BinaryDataWriter) writeString(s string) {
	runes := []rune(s)
	charCount := int32(len(runes))

	w.writeByte(1) // charSizeFlag = 1 (16-bit)
	w.writeInt32(charCount)
	u16 := utf16.Encode(runes)
	for _, v := range u16 {
		w.writeUint16(v)
	}
}

// writeTypeEntry writes a TypeInfo to the stream.
// Matches C# format: TypeName writes byte(47) + int32_id + string.
func (w *BinaryDataWriter) writeTypeEntry(ti *TypeInfo) {
	if ti == nil {
		w.writeByte(byte(UnnamedNull))
		return
	}
	if ti.IsTypeID {
		w.writeByte(byte(TypeID))
		w.writeInt32(ti.TypeIDValue)
	} else {
		w.writeByte(byte(TypeName))
		w.writeInt32(ti.TypeIDValue)
		w.writeString(ti.TypeName)
	}
}

// ---- High-level tree writer ----

// WriteTree writes a Node tree to binary Odin format.
func (w *BinaryDataWriter) WriteTree(node *Node) error {
	return w.writeNode(node)
}

func (w *BinaryDataWriter) writeNode(node *Node) error {
	bt := node.BinType

	// Write entry type byte.
	w.writeByte(byte(bt))

	// Write name for named entries.
	if bt.IsNamed() {
		w.writeString(node.Name)
	}

	switch bt {
	case NamedStartOfReferenceNode, UnnamedStartOfReferenceNode:
		w.writeTypeEntry(node.TypeInfo)
		w.writeInt32(node.NodeID)
		for _, child := range node.Children {
			if err := w.writeNode(child); err != nil {
				return err
			}
		}
		w.writeByte(byte(EndOfNode))

	case NamedStartOfStructNode, UnnamedStartOfStructNode:
		w.writeTypeEntry(node.TypeInfo)
		for _, child := range node.Children {
			if err := w.writeNode(child); err != nil {
				return err
			}
		}
		w.writeByte(byte(EndOfNode))

	case StartOfArray:
		w.writeInt64(node.ArrayLength)
		for _, child := range node.Children {
			if err := w.writeNode(child); err != nil {
				return err
			}
		}
		w.writeByte(byte(EndOfArray))

	case EndOfNode, EndOfArray, BinaryEndOfStream:
		// Already written the entry byte above.

	case PrimitiveArray:
		pai := node.PrimArrayInfo
		if pai == nil {
			return fmt.Errorf("PrimitiveArray node missing PrimArrayInfo")
		}
		raw := extractBytes(node.Value)
		if pai.BytesPerElement == 1 {
			pai.ElementCount = int32(len(raw))
		}
		w.writeInt32(pai.ElementCount)
		w.writeInt32(pai.BytesPerElement)
		w.writeBytes(raw)

	case NamedInternalReference, UnnamedInternalReference:
		w.writeInt32(toInt32(node.Value))

	case NamedExternalReferenceByIndex, UnnamedExternalReferenceByIndex:
		w.writeInt32(toInt32(node.Value))

	case NamedExternalReferenceByGuid, UnnamedExternalReferenceByGuid:
		raw := extractBytes(node.Value)
		w.writeBytes(raw)

	case NamedExternalReferenceByString, UnnamedExternalReferenceByString:
		w.writeString(toString(node.Value))

	case NamedSByte, UnnamedSByte:
		w.writeByte(byte(toInt8(node.Value)))

	case NamedByte, UnnamedByte:
		w.writeByte(toUint8(node.Value))

	case NamedShort, UnnamedShort:
		w.writeInt16(toInt16(node.Value))

	case NamedUShort, UnnamedUShort:
		w.writeUint16(toUint16(node.Value))

	case NamedInt, UnnamedInt:
		w.writeInt32(toInt32(node.Value))

	case NamedUInt, UnnamedUInt:
		w.writeUint32(toUint32(node.Value))

	case NamedLong, UnnamedLong:
		w.writeInt64(toInt64(node.Value))

	case NamedULong, UnnamedULong:
		w.writeUint64(toUint64(node.Value))

	case NamedFloat, UnnamedFloat:
		w.writeFloat32(toFloat32(node.Value))

	case NamedDouble, UnnamedDouble:
		w.writeFloat64(toFloat64(node.Value))

	case NamedDecimal, UnnamedDecimal:
		raw := extractBytes(node.Value)
		w.writeBytes(raw)

	case NamedChar, UnnamedChar:
		s := toString(node.Value)
		if len(s) > 0 {
			runes := []rune(s)
			w.writeUint16(uint16(runes[0]))
		} else {
			w.writeUint16(0)
		}

	case NamedString, UnnamedString:
		w.writeString(toString(node.Value))

	case NamedGuid, UnnamedGuid:
		raw := extractBytes(node.Value)
		w.writeBytes(raw)

	case NamedBoolean, UnnamedBoolean:
		if toBool(node.Value) {
			w.writeByte(1)
		} else {
			w.writeByte(0)
		}

	case NamedNull, UnnamedNull:
		// No payload.

	default:
		return fmt.Errorf("unknown entry type %d in node", bt)
	}

	return nil
}

// ---- type conversion helpers for JSON round-trip ----

func extractBytes(v interface{}) []byte {
	switch val := v.(type) {
	case ByteSliceWrapper:
		return val.Data
	case []byte:
		return val
	case *Node:
		writer := NewBinaryDataWriter()
		if err := writer.WriteTree(val); err == nil {
			bytes := writer.Bytes()
			fmt.Printf("Serialized sub-tree into %d bytes\n", len(bytes))
			return bytes
		} else {
			fmt.Printf("Failed to serialize sub-tree: %v\n", err)
		}
		return nil
	default:
		return nil
	}
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func toBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	default:
		return false
	}
}

func toInt8(v interface{}) int8 {
	switch val := v.(type) {
	case int8:
		return val
	case float64:
		return int8(val)
	default:
		return 0
	}
}

func toUint8(v interface{}) byte {
	switch val := v.(type) {
	case byte:
		return val
	case float64:
		return byte(val)
	default:
		return 0
	}
}

func toInt16(v interface{}) int16 {
	switch val := v.(type) {
	case int16:
		return val
	case float64:
		return int16(val)
	default:
		return 0
	}
}

func toUint16(v interface{}) uint16 {
	switch val := v.(type) {
	case uint16:
		return val
	case float64:
		return uint16(val)
	default:
		return 0
	}
}

func toInt32(v interface{}) int32 {
	switch val := v.(type) {
	case int32:
		return val
	case float64:
		return int32(val)
	default:
		return 0
	}
}

func toUint32(v interface{}) uint32 {
	switch val := v.(type) {
	case uint32:
		return val
	case float64:
		return uint32(val)
	default:
		return 0
	}
}

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case float64:
		return int64(val)
	default:
		return 0
	}
}

func toUint64(v interface{}) uint64 {
	switch val := v.(type) {
	case uint64:
		return val
	case float64:
		return uint64(val)
	default:
		return 0
	}
}

func toFloat32(v interface{}) float32 {
	switch val := v.(type) {
	case float32:
		return val
	case float64:
		return float32(val)
	default:
		return 0
	}
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	default:
		return 0
	}
}
