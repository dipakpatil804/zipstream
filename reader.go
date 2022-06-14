/*
* In this file we refer to ZIP format please refer below link for more info,
* https://en.wikipedia.org/wiki/ZIP_(file_format)
 */

package zipstream

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"fmt"
	"io"
	"log"
)

type ZipStream struct {
	Reader   io.Reader
	previous []byte
}

// Workflow:
// Zip has format like below:
// Header...Data...Header...Data...Descriptor...CentralDirHeader...CentralDirDescriptor
// High level workflow:
// Read Local File header
// Read data until Descriptor or next Local File Header or Central Directory Header
// if Compression method is set to 8 use deflate Reader to read data.
// Next will return the io.EOF when the zip end is reached.
// Read will return the io.EOF when current file end is reached.

// Next reads the local file header of the next file and returns the Reader to read data.
// if compression method is 8, then deflate Reader will be returned.
func (z *ZipStream) Next() (*LocalFileHeader, io.Reader, error) {
	var reader io.Reader
	var totalDataRead = 0
	if z.previous != nil && len(z.previous) != 0 {
		reader = AppendReader(bytes.NewReader(z.previous), z.Reader)
	} else {
		reader = z.Reader
	}

	buff := make([]byte, 4)
	n, err := reader.Read(buff)
	if err != nil && err != io.EOF {
		return nil, nil, err
	}

	if n != len(buff) {
		return nil, nil, fmt.Errorf("failed to read header signature, "+
			"expected %d bytes got %d\n", len(buff), n)
	}

	totalDataRead += n
	if binary.LittleEndian.Uint32(buff) == directoryHeaderSignature {
		return nil, nil, io.EOF
	}

	if binary.LittleEndian.Uint32(buff) == localFileHeaderSignature {
		lfh, tdr, err := z.readLocalFileHeader(reader)
		if err != nil {
			return nil, nil, err
		}
		totalDataRead += tdr
		if tdr >= len(z.previous) {
			z.previous = nil
		} else {
			z.previous = z.previous[totalDataRead:]
		}

		switch lfh.CMethod {
		case 8:
			return lfh, flate.NewReader(z), nil
		case 0:
			return lfh, z, nil
		default:
			log.Panicf("unknown compression method: %d", lfh.CMethod)
		}
	}
	return nil, nil, fmt.Errorf("invalid signature: %x", binary.LittleEndian.Uint32(buff))
}

// Read reads the data from the previous remaining data and from Reader to read next data.
func (z *ZipStream) Read(p []byte) (n int, err error) {
	var reader io.Reader
	if z.previous != nil && len(z.previous) != 0 {
		reader = AppendReader(bytes.NewReader(z.previous), z.Reader)
	} else {
		reader = z.Reader
	}
	n, err = reader.Read(p)
	p = p[:n]

	if n >= len(z.previous) {
		z.previous = nil
	}

	for i := 0; i < len(p)-4; i++ {
		if binary.LittleEndian.Uint32(p[i:i+4]) == localFileHeaderSignature ||
			binary.LittleEndian.Uint32(p[i:i+4]) == directoryHeaderSignature ||
			binary.LittleEndian.Uint32(p[i:i+4]) == dataDescriptorSignature {
			discardDataDescriptor := binary.LittleEndian.Uint32(p[i:i+4]) == dataDescriptorSignature
			if z.previous == nil {
				z.previous = p[i:]
				p = p[:i]
			} else {
				z.previous = append(p[i:], z.previous[len(p):]...)
				p = p[:i]
			}

			if discardDataDescriptor {
				if z.previous != nil && len(z.previous) != 0 {
					reader = AppendReader(bytes.NewReader(z.previous), z.Reader)
				} else {
					reader = z.Reader
				}
				buff := make([]byte, 40)
				n, err = reader.Read(buff)
				buff = buff[:n]
				if n >= len(z.previous) {
					z.previous = nil
				}
				var found bool
				for i := 0; i < len(buff)-4; i++ {
					if binary.LittleEndian.Uint32(buff[i:i+4]) == localFileHeaderSignature ||
						binary.LittleEndian.Uint32(buff[i:i+4]) == directoryHeaderSignature {
						if z.previous == nil {
							z.previous = buff[i:]
						} else {
							z.previous = append(buff[i:], z.previous[len(buff):]...)
						}
						found = true
						break
					}
				}
				if !found {
					return 0, fmt.Errorf("invalid read")
				}
			}
			return len(p), io.EOF
		}
	}
	if z.previous == nil {
		z.previous = p[len(p)-4:]
		p = p[:len(p)-4]
	} else {
		z.previous = append(p[len(p)-4:], z.previous[len(p)-4:]...)
		p = p[:len(p)-4]
	}
	return len(p), nil
}

func (z *ZipStream) readLocalFileHeader(reader io.Reader) (*LocalFileHeader, int, error) {
	totalDataRead := 0
	buff := make([]byte, 26)
	n, err := reader.Read(buff)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read local file header, error : %s", err.Error())
	}
	if n != len(buff) {
		return nil, 0, fmt.Errorf("failed to read local file header, "+
			"expected %d bytes got %d\n", len(buff), n)
	}
	totalDataRead += n

	r := LittleEndianReader{bytes.NewReader(buff)}
	localFileHeader := new(LocalFileHeader)
	localFileHeader.FileHeader = localFileHeaderSignature
	localFileHeader.Version = r.uint16()
	localFileHeader.GeneralBits = r.uint16()
	localFileHeader.CMethod = r.uint16()
	localFileHeader.ModifiedTime = r.uint16()
	localFileHeader.ModifiedDate = r.uint16()
	localFileHeader.CRC32 = r.uint32()
	localFileHeader.CompressedSize = r.uint32()
	localFileHeader.UnCompressedSize = r.uint32()
	localFileHeader.FileNameLength = r.uint16()
	localFileHeader.ExtrasLength = r.uint16()

	if localFileHeader.FileNameLength != 0 {
		buff = make([]byte, localFileHeader.FileNameLength)
		n, err = reader.Read(buff)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to read file name, error : %s", err.Error())
		}
		if n != len(buff) {
			return nil, 0, fmt.Errorf("failed to read file name, "+
				"expected %d bytes got %d\n", len(buff), n)
		}
		localFileHeader.FileName = string(buff)
		totalDataRead += n
	}

	if localFileHeader.ExtrasLength != 0 {
		buff = make([]byte, localFileHeader.ExtrasLength)
		n, err = reader.Read(buff)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to read extras, error : %s", err.Error())
		}
		if n != len(buff) {
			return nil, 0, fmt.Errorf("failed to read extras, "+
				"expected %d bytes got %d\n", len(buff), n)
		}
		localFileHeader.Extras = string(buff)
		totalDataRead += n
	}

	if localFileHeader.FileName[len(localFileHeader.FileName)-1] == '/' {
		localFileHeader.IsDirectory = true
	}
	return localFileHeader, totalDataRead, nil
}
