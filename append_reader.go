package zipstream

import "io"

// AppendReader is similar to io.MultiReader, Unlike io.MultiReader it appends
// the data from second reader in same buffer without making extra Read call
func AppendReader(readers ...io.Reader) io.Reader {
	r := make([]io.Reader, len(readers))
	copy(r, readers)
	return &appendReader{r}
}

type appendReader struct {
	readers []io.Reader
}

func (m *appendReader) Read(p []byte) (n int, err error) {
	dataToRead := len(p)
	dataRead := 0
	for _, r := range m.readers {
		var buff = make([]byte, dataToRead-dataRead)
		if r != nil {
			n, err = r.Read(buff)
			if err != nil && err != io.EOF {
				return 0, err
			}
			for i := 0; i < n; i++ {
				p[dataRead+i] = buff[i]
			}
			dataRead += n
			if dataRead == dataToRead {
				return dataRead, nil
			}
			if err == io.EOF {
				r = nil
				err = nil
				continue
			}
		}
	}
	return dataRead, io.EOF
}
