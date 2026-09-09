# skills/go-documentation/scripts/check-docs-ast.go

- missingDoc · struct · L18-L23 — missingDoc
- parseError · struct · L25-L28 — parseError
- parsedFile · struct · L30-L33 — parsedFile
- packageInfo · struct · L35-L40 — packageInfo
- usage · function · L42-L55 — func usage()
- main · function · L57-L199 — func main()
- options · struct · L201-L208 — options
- parseArgs · function · L210-L254 — func parseArgs(args []string) (options, error)
- parseNonNegativeInt · function · L256-L268 — func parseNonNegativeInt(s string) (int, error)
- findGoFiles · function · L270-L293 — func findGoFiles(target string) ([]string, error)
- walkGoFiles · function · L295-L315 — func walkGoFiles(root string) ([]string, error)
- packageKey · function · L317-L319 — func packageKey(path, name string) string
- findMissingDeclDocs · function · L321-L373 — func findMissingDeclDocs(fset *token.FileSet, pf parsedFile, strict bool) []missingDoc
