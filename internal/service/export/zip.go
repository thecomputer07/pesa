package export

import (
	"archive/zip"
	"bytes"
	"io"
)

// BundleFile is one entry in the output zip.
type BundleFile struct {
	Name string
	Body []byte
}

// WriteBundle streams a zip with deflate compression of each provided file.
// Returns the number of bytes written to w.
func WriteBundle(w io.Writer, parts []BundleFile) (int64, error) {
	var cw countingWriter
	cw.w = w
	zw := zip.NewWriter(&cw)
	for _, p := range parts {
		fh := &zip.FileHeader{Name: p.Name, Method: zip.Deflate}
		fw, err := zw.CreateHeader(fh)
		if err != nil {
			return cw.n, err
		}
		if _, err := fw.Write(p.Body); err != nil {
			return cw.n, err
		}
	}
	if err := zw.Close(); err != nil {
		return cw.n, err
	}
	return cw.n, nil
}

// BundleToBytes is a convenience for tests.
func BundleToBytes(parts []BundleFile) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := WriteBundle(&buf, parts); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type countingWriter struct {
	w io.Writer
	n int64
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)
	return n, err
}

// PartsFromTables turns CollectTables output into the byte-payload slice
// consumed by WriteBundle.
func PartsFromTables(account []byte, tables []TablePart, readme []byte) ([]BundleFile, error) {
	out := make([]BundleFile, 0, len(tables)+2)
	out = append(out, BundleFile{Name: "account.json", Body: account})
	for _, t := range tables {
		var buf bytes.Buffer
		if err := WriteCSV(&buf, t.Header, t.Rows); err != nil {
			return nil, err
		}
		out = append(out, BundleFile{Name: t.Name, Body: buf.Bytes()})
	}
	out = append(out, BundleFile{Name: "README.md", Body: readme})
	return out, nil
}
