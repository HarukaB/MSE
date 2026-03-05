package odin

import (
	"encoding/binary"
	"fmt"
	"io"
	"unicode/utf16"
)

// BinaryDataReader reads Odin Serializer binary format.
type BinaryDataReader struct {
	data []byte
	pos  int
	// types maps type IDs to type name strings, built during reading.
	types map[int32]string
}

// NewBinaryDataReader creates a reader from raw bytes.
func NewBinaryDataReader(data []byte) *BinaryDataReader {
	return &BinaryDataReader{
		data:  data,
		pos:   0,
		types: make(map[int32]string),
	}
}

// Remaining returns the number of unread bytes.
func (r *BinaryDataReader) Remaining() int {
	return len(r.data) - r.pos
}

// ---- low-level read helpers ----

func (r *BinaryDataReader) readByte() (byte, error) {
	if r.pos >= len(r.data) {
		return 0, io.ErrUnexpectedEOF
	}
	b := r.data[r.pos]
	r.pos++
	return b, nil
}

func (r *BinaryDataReader) readBytes(n int) ([]byte, error) {
	if r.pos+n > len(r.data) {
		return nil, io.ErrUnexpectedEOF
	}
	b := make([]byte, n)
	copy(b, r.data[r.pos:r.pos+n])
	r.pos += n
	return b, nil
}

func (r *BinaryDataReader) readInt32() (int32, error) {
	if r.pos+4 > len(r.data) {
		return 0, io.ErrUnexpectedEOF
	}
	v := int32(binary.LittleEndian.Uint32(r.data[r.pos:]))
	r.pos += 4
	return v, nil
}

func (r *BinaryDataReader) readUint32() (uint32, error) {
	if r.pos+4 > len(r.data) {
		return 0, io.ErrUnexpectedEOF
	}
	v := binary.LittleEndian.Uint32(r.data[r.pos:])
	r.pos += 4
	return v, nil
}

func (r *BinaryDataReader) readInt64() (int64, error) {
	if r.pos+8 > len(r.data) {
		return 0, io.ErrUnexpectedEOF
	}
	v := int64(binary.LittleEndian.Uint64(r.data[r.pos:]))
	r.pos += 8
	return v, nil
}

func (r *BinaryDataReader) readUint64() (uint64, error) {
	if r.pos+8 > len(r.data) {
		return 0, io.ErrUnexpectedEOF
	}
	v := binary.LittleEndian.Uint64(r.data[r.pos:])
	r.pos += 8
	return v, nil
}

func (r *BinaryDataReader) readFloat32() (float32, error) {
	if r.pos+4 > len(r.data) {
		return 0, io.ErrUnexpectedEOF
	}
	bits := binary.LittleEndian.Uint32(r.data[r.pos:])
	r.pos += 4
	return float32FromBits(bits), nil
}

func (r *BinaryDataReader) readFloat64() (float64, error) {
	if r.pos+8 > len(r.data) {
		return 0, io.ErrUnexpectedEOF
	}
	bits := binary.LittleEndian.Uint64(r.data[r.pos:])
	r.pos += 8
	return float64FromBits(bits), nil
}

// readString reads an Odin-encoded string: 1 byte charSizeFlag + int32 charCount + data.
func (r *BinaryDataReader) readString() (string, error) {
	charSizeFlag, err := r.readByte()
	if err != nil {
		return "", err
	}

	charCount, err := r.readInt32()
	if err != nil {
		return "", err
	}

	if charCount < 0 {
		return "", fmt.Errorf("negative string char count: %d", charCount)
	}
	if charCount == 0 {
		return "", nil
	}

	if charSizeFlag == 0 {
		// 8-bit characters — each byte is one char (low byte of UTF-16).
		if r.pos+int(charCount) > len(r.data) {
			return "", io.ErrUnexpectedEOF
		}
		runes := make([]rune, charCount)
		for i := int32(0); i < charCount; i++ {
			runes[i] = rune(r.data[r.pos+int(i)])
		}
		r.pos += int(charCount)
		return string(runes), nil
	}

	// 16-bit characters — UTF-16LE, charCount * 2 bytes.
	byteCount := int(charCount) * 2
	if r.pos+byteCount > len(r.data) {
		return "", io.ErrUnexpectedEOF
	}
	u16 := make([]uint16, charCount)
	for i := int32(0); i < charCount; i++ {
		u16[i] = binary.LittleEndian.Uint16(r.data[r.pos+int(i)*2:])
	}
	r.pos += byteCount
	return string(utf16.Decode(u16)), nil
}

// readTypeEntry reads a TypeName or TypeID entry, registering new type IDs.
// Matches C# ReadTypeEntry: consumes the entry byte, then:
// - TypeName(47): int32 id, then string name
// - TypeID(48): int32 id
// - UnnamedNull(46): null type
// - anything else: error (but we treat as no type for robustness)
func (r *BinaryDataReader) readTypeEntry() (*TypeInfo, error) {
	if r.pos >= len(r.data) {
		return nil, nil
	}

	peek := BinaryEntryType(r.data[r.pos])

	switch peek {
	case TypeName:
		r.pos++ // consume the TypeName byte
		id, err := r.readInt32()
		if err != nil {
			return nil, fmt.Errorf("reading type name id: %w", err)
		}
		name, err := r.readString()
		if err != nil {
			return nil, fmt.Errorf("reading type name string: %w", err)
		}
		r.types[id] = name
		return &TypeInfo{TypeName: name, TypeIDValue: id}, nil

	case TypeID:
		r.pos++ // consume the TypeID byte
		id, err := r.readInt32()
		if err != nil {
			return nil, fmt.Errorf("reading type id: %w", err)
		}
		name := r.types[id] // look up previously registered name
		return &TypeInfo{IsTypeID: true, TypeName: name, TypeIDValue: id}, nil

	case UnnamedNull:
		r.pos++ // consume the UnnamedNull byte — means null type
		return nil, nil

	default:
		// No type entry — valid for untyped nodes.
		return nil, nil
	}
}

// ---- High-level tree reader ----

// ReadTree reads the entire Odin binary stream into a Node tree.
func (r *BinaryDataReader) ReadTree() (*Node, error) {
	root := &Node{BinType: 0, Name: "$root"}
	for r.pos < len(r.data) {
		child, err := r.readEntry()
		if err != nil {
			return nil, fmt.Errorf("at offset 0x%x: %w", r.pos, err)
		}
		if child == nil {
			break // EndOfStream
		}
		root.Children = append(root.Children, child)
	}
	// If root has exactly one child, unwrap it.
	if len(root.Children) == 1 {
		return root.Children[0], nil
	}
	return root, nil
}

// readEntry reads a single entry and its content from the stream.
func (r *BinaryDataReader) readEntry() (*Node, error) {
	if r.pos >= len(r.data) {
		return nil, nil
	}

	entryByte, err := r.readByte()
	if err != nil {
		return nil, err
	}
	bt := BinaryEntryType(entryByte)
	node := &Node{BinType: bt}

	// Read name for named entries.
	if bt.IsNamed() {
		name, err := r.readString()
		if err != nil {
			return nil, fmt.Errorf("reading name for %d: %w", bt, err)
		}
		node.Name = name
	}

	switch bt {
	// ---- Node start ----
	case NamedStartOfReferenceNode, UnnamedStartOfReferenceNode:
		ti, err := r.readTypeEntry()
		if err != nil {
			return nil, err
		}
		node.TypeInfo = ti
		id, err := r.readInt32()
		if err != nil {
			return nil, err
		}
		node.NodeID = id
		// Read children until EndOfNode.
		if err := r.readNodeChildren(node); err != nil {
			return nil, err
		}

	case NamedStartOfStructNode, UnnamedStartOfStructNode:
		ti, err := r.readTypeEntry()
		if err != nil {
			return nil, err
		}
		node.TypeInfo = ti
		node.NodeID = -1
		// Read children until EndOfNode.
		if err := r.readNodeChildren(node); err != nil {
			return nil, err
		}

	// ---- Array ----
	case StartOfArray:
		length, err := r.readInt64()
		if err != nil {
			return nil, err
		}
		node.ArrayLength = length
		// Read children until EndOfArray.
		if err := r.readArrayChildren(node); err != nil {
			return nil, err
		}

	// ---- Structure markers ----
	case EndOfNode, EndOfArray:
		// These are handled by the parent reader, but if we see them at top level,
		// return them as-is so the parent can detect them.
		return node, nil

	case BinaryEndOfStream:
		return nil, nil // signal end

	// ---- Primitive Array ----
	case PrimitiveArray:
		elemCount, err := r.readInt32()
		if err != nil {
			return nil, err
		}
		bytesPerElem, err := r.readInt32()
		if err != nil {
			return nil, err
		}
		node.PrimArrayInfo = &PrimArrayInfo{
			ElementCount:    elemCount,
			BytesPerElement: bytesPerElem,
		}
		totalBytes := int(elemCount) * int(bytesPerElem)
		raw, err := r.readBytes(totalBytes)
		if err != nil {
			return nil, err
		}

		// Try to parse the inner byte array as a standalone Odin Tree
		subReader := NewBinaryDataReader(raw)
		subTree, subErr := subReader.ReadTree()
		if subErr == nil && subTree != nil && subReader.Remaining() == 0 {
			node.Value = subTree
		} else {
			node.Value = ByteSliceWrapper{Data: raw}
		}

	// ---- Internal / External References ----
	case NamedInternalReference, UnnamedInternalReference:
		v, err := r.readInt32()
		if err != nil {
			return nil, err
		}
		node.Value = v

	case NamedExternalReferenceByIndex, UnnamedExternalReferenceByIndex:
		v, err := r.readInt32()
		if err != nil {
			return nil, err
		}
		node.Value = v

	case NamedExternalReferenceByGuid, UnnamedExternalReferenceByGuid:
		raw, err := r.readBytes(16)
		if err != nil {
			return nil, err
		}
		node.Value = ByteSliceWrapper{Data: raw}

	case NamedExternalReferenceByString, UnnamedExternalReferenceByString:
		v, err := r.readString()
		if err != nil {
			return nil, err
		}
		node.Value = v

	// ---- Primitive types ----
	case NamedSByte, UnnamedSByte:
		b, err := r.readByte()
		if err != nil {
			return nil, err
		}
		node.Value = int8(b)

	case NamedByte, UnnamedByte:
		b, err := r.readByte()
		if err != nil {
			return nil, err
		}
		node.Value = b

	case NamedShort, UnnamedShort:
		if r.pos+2 > len(r.data) {
			return nil, io.ErrUnexpectedEOF
		}
		v := int16(binary.LittleEndian.Uint16(r.data[r.pos:]))
		r.pos += 2
		node.Value = v

	case NamedUShort, UnnamedUShort:
		if r.pos+2 > len(r.data) {
			return nil, io.ErrUnexpectedEOF
		}
		v := binary.LittleEndian.Uint16(r.data[r.pos:])
		r.pos += 2
		node.Value = v

	case NamedInt, UnnamedInt:
		v, err := r.readInt32()
		if err != nil {
			return nil, err
		}
		node.Value = v

	case NamedUInt, UnnamedUInt:
		v, err := r.readUint32()
		if err != nil {
			return nil, err
		}
		node.Value = v

	case NamedLong, UnnamedLong:
		v, err := r.readInt64()
		if err != nil {
			return nil, err
		}
		node.Value = v

	case NamedULong, UnnamedULong:
		v, err := r.readUint64()
		if err != nil {
			return nil, err
		}
		node.Value = v

	case NamedFloat, UnnamedFloat:
		v, err := r.readFloat32()
		if err != nil {
			return nil, err
		}
		node.Value = v

	case NamedDouble, UnnamedDouble:
		v, err := r.readFloat64()
		if err != nil {
			return nil, err
		}
		node.Value = v

	case NamedDecimal, UnnamedDecimal:
		raw, err := r.readBytes(16)
		if err != nil {
			return nil, err
		}
		node.Value = ByteSliceWrapper{Data: raw}

	case NamedChar, UnnamedChar:
		if r.pos+2 > len(r.data) {
			return nil, io.ErrUnexpectedEOF
		}
		v := binary.LittleEndian.Uint16(r.data[r.pos:])
		r.pos += 2
		node.Value = string(rune(v))

	case NamedString, UnnamedString:
		v, err := r.readString()
		if err != nil {
			return nil, err
		}
		node.Value = v

	case NamedGuid, UnnamedGuid:
		raw, err := r.readBytes(16)
		if err != nil {
			return nil, err
		}
		node.Value = ByteSliceWrapper{Data: raw}

	case NamedBoolean, UnnamedBoolean:
		b, err := r.readByte()
		if err != nil {
			return nil, err
		}
		node.Value = b != 0

	case NamedNull, UnnamedNull:
		// No payload.
		node.Value = nil

	default:
		return nil, fmt.Errorf("unknown entry type %d at offset 0x%x", bt, r.pos-1)
	}

	return node, nil
}

// readNodeChildren reads child entries until EndOfNode.
func (r *BinaryDataReader) readNodeChildren(parent *Node) error {
	for r.pos < len(r.data) {
		// Peek at next byte.
		if BinaryEntryType(r.data[r.pos]) == EndOfNode {
			r.pos++ // consume EndOfNode
			return nil
		}
		child, err := r.readEntry()
		if err != nil {
			return err
		}
		if child == nil {
			return fmt.Errorf("unexpected end of stream inside node")
		}
		if child.BinType == EndOfNode {
			return nil
		}
		parent.Children = append(parent.Children, child)
	}
	return fmt.Errorf("unexpected EOF: missing EndOfNode")
}

// readArrayChildren reads child entries until EndOfArray.
func (r *BinaryDataReader) readArrayChildren(parent *Node) error {
	for r.pos < len(r.data) {
		if BinaryEntryType(r.data[r.pos]) == EndOfArray {
			r.pos++ // consume EndOfArray
			return nil
		}
		child, err := r.readEntry()
		if err != nil {
			return err
		}
		if child == nil {
			return fmt.Errorf("unexpected end of stream inside array")
		}
		if child.BinType == EndOfArray {
			return nil
		}
		parent.Children = append(parent.Children, child)
	}
	return fmt.Errorf("unexpected EOF: missing EndOfArray")
}
