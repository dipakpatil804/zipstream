package zipstream

const (
	localFileHeaderSignature = 0x04034b50
	directoryHeaderSignature = 0x02014b50
	dataDescriptorSignature  = 0x08074b50
)

type LocalFileHeader struct {
	FileHeader       uint32
	Version          uint16
	GeneralBits      uint16
	CMethod          uint16
	ModifiedTime     uint16
	ModifiedDate     uint16
	CRC32            uint32
	CompressedSize   uint32
	UnCompressedSize uint32
	FileNameLength   uint16
	ExtrasLength     uint16
	FileName         string
	Extras           string
	IsDirectory      bool
}
