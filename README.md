## About this project

This project is a ZIP file reader which gives a decompressed stream of ZIP.
This project is implemented with the ZIP format specification given in below wiki page.
https://en.wikipedia.org/wiki/ZIP_(file_format)

##Types
###ZipStream
```
type ZipStream struct {
    Reader   io.Reader
}
```
**ZipStream** accepts an io.Reader to ZIP file as input.

##Functions
```
func Next() (*LocalFileHeader, io.Reader, error)
```
Next() function will return the LocalFileHeader object which has file name and other details.
It also returns the io.Reader to the next file to read. Function will return io.EOF when there 
are no more files in ZIP.

####Note: If you don't want to read current file, then you need to discard the content of current file. 

**io.Copy(ioutil.Discard, reader)**

```
func Read(p []byte) (n int, err error)
```
Read function should only be called on the io.Reader return by Next() function.
It will fill the **p []byte** with the decompressed bytes of current file.
## Installation 
```
go get -v "github.com/dipakpatil804/zipstream"
```

## Sample Example
```
    f, err := os.Open("~/sample.zip")
	if err != nil {
		fmt.Printf("Failed to open file.\n")
		return
	}

	z := zipstream.ZipStream{Reader: f}
	for {
		lfh, r, err := z.Next()
		if err == io.EOF {
			fmt.Printf("Done Reading Zip\n")
			return
		}
		fmt.Printf("Reading file: %s\n", lfh.FileName)
		
		if lfh.IsDirectory {
			continue
		}

		io.Copy(ioutil.Discard, r)
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Printf("Error: %s", err.Error())
			return
		}
	}
```