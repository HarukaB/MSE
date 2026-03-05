package odin

// BinaryEntryType represents the low-level binary entry types used in Odin Serializer's binary format.
// IMPORTANT: These values match the GAME's version (from dump.cs), which differs from the
// open-source Odin Serializer. ExternalReferenceByString is moved to 50-51 (end),
// and all types from SByte onwards are shifted down by 2.
type BinaryEntryType byte

const (
	InvalidEntry                     BinaryEntryType = 0
	NamedStartOfReferenceNode        BinaryEntryType = 1
	UnnamedStartOfReferenceNode      BinaryEntryType = 2
	NamedStartOfStructNode           BinaryEntryType = 3
	UnnamedStartOfStructNode         BinaryEntryType = 4
	EndOfNode                        BinaryEntryType = 5
	StartOfArray                     BinaryEntryType = 6
	EndOfArray                       BinaryEntryType = 7
	PrimitiveArray                   BinaryEntryType = 8
	NamedInternalReference           BinaryEntryType = 9
	UnnamedInternalReference         BinaryEntryType = 10
	NamedExternalReferenceByIndex    BinaryEntryType = 11
	UnnamedExternalReferenceByIndex  BinaryEntryType = 12
	NamedExternalReferenceByGuid     BinaryEntryType = 13
	UnnamedExternalReferenceByGuid   BinaryEntryType = 14
	NamedSByte                       BinaryEntryType = 15
	UnnamedSByte                     BinaryEntryType = 16
	NamedByte                        BinaryEntryType = 17
	UnnamedByte                      BinaryEntryType = 18
	NamedShort                       BinaryEntryType = 19
	UnnamedShort                     BinaryEntryType = 20
	NamedUShort                      BinaryEntryType = 21
	UnnamedUShort                    BinaryEntryType = 22
	NamedInt                         BinaryEntryType = 23
	UnnamedInt                       BinaryEntryType = 24
	NamedUInt                        BinaryEntryType = 25
	UnnamedUInt                      BinaryEntryType = 26
	NamedLong                        BinaryEntryType = 27
	UnnamedLong                      BinaryEntryType = 28
	NamedULong                       BinaryEntryType = 29
	UnnamedULong                     BinaryEntryType = 30
	NamedFloat                       BinaryEntryType = 31
	UnnamedFloat                     BinaryEntryType = 32
	NamedDouble                      BinaryEntryType = 33
	UnnamedDouble                    BinaryEntryType = 34
	NamedDecimal                     BinaryEntryType = 35
	UnnamedDecimal                   BinaryEntryType = 36
	NamedChar                        BinaryEntryType = 37
	UnnamedChar                      BinaryEntryType = 38
	NamedString                      BinaryEntryType = 39
	UnnamedString                    BinaryEntryType = 40
	NamedGuid                        BinaryEntryType = 41
	UnnamedGuid                      BinaryEntryType = 42
	NamedBoolean                     BinaryEntryType = 43
	UnnamedBoolean                   BinaryEntryType = 44
	NamedNull                        BinaryEntryType = 45
	UnnamedNull                      BinaryEntryType = 46
	TypeName                         BinaryEntryType = 47
	TypeID                           BinaryEntryType = 48
	BinaryEndOfStream                BinaryEntryType = 49
	NamedExternalReferenceByString   BinaryEntryType = 50
	UnnamedExternalReferenceByString BinaryEntryType = 51
)

// IsNamed returns true if this entry type carries a name.
func (b BinaryEntryType) IsNamed() bool {
	switch b {
	case NamedStartOfReferenceNode, NamedStartOfStructNode,
		NamedInternalReference,
		NamedExternalReferenceByIndex, NamedExternalReferenceByGuid, NamedExternalReferenceByString,
		NamedSByte, NamedByte, NamedShort, NamedUShort,
		NamedInt, NamedUInt, NamedLong, NamedULong,
		NamedFloat, NamedDouble, NamedDecimal,
		NamedChar, NamedString, NamedGuid, NamedBoolean, NamedNull:
		return true
	}
	return false
}
