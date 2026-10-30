# Archive Entries

> Sources: `archive/tar`, `archive/zip`, and `os` (`Root`) package docs
> Authority: normative for the extraction checks
> Minimum Go: 1.25 for `os.Root.MkdirAll`
> Last verified: 2026-10-30

An entry's name, type, and size are all input. Create parents with
`root.MkdirAll`, never `os.MkdirAll(filepath.Join(dest, dir))`, which runs
outside the root; create no link an entry names; and cap the total size:

```go
func extract(root *os.Root, tr *tar.Reader, limit int64) error {
    for {
        hdr, err := tr.Next()
        if errors.Is(err, io.EOF) {
            return nil
        }
        if err != nil {
            return err
        }
        switch hdr.Typeflag {
        case tar.TypeReg: // extracted below
        case tar.TypeSymlink, tar.TypeLink:
            return fmt.Errorf("tar entry %q: links are not extracted", hdr.Name)
        default:
            continue // directories come from MkdirAll; devices and pax headers are skipped
        }
        if limit -= hdr.Size; limit < 0 {
            return errors.New("archive exceeds the size limit")
        }
        if err := root.MkdirAll(path.Dir(hdr.Name), 0o750); err != nil {
            return err // "../x" and "/x" escape the root and fail here
        }
        f, err := root.Create(hdr.Name)
        if err != nil {
            return err
        }
        _, err = io.Copy(f, tr)
        if err = errors.Join(err, f.Close()); err != nil {
            return err
        }
    }
}
```

An `archive/zip` entry takes the same three checks: `f.Mode().IsRegular()`,
`f.UncompressedSize64` against the cap, and the name through the root.
