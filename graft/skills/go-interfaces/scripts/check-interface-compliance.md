# skills/go-interfaces/scripts/check-interface-compliance.go

- ifaceInfo · struct · L20-L27 — ifaceInfo
- result · struct · L29-L36 — result
- sourceFile · struct · L38-L42 — sourceFile
- packageGroup · struct · L44-L49 — packageGroup
- options · struct · L51-L58 — options
- usage · function · L60-L73 — func usage()
- main · function · L75-L192 — func main()
- emit · function · L194-L233 — func emit(out result, jsonOutput bool)
- analyzePackage · function · L235-L341 — func analyzePackage(group packageGroup) (map[string]bool, []ifaceInfo, map[string]bool, error)
- typeBelongsToScannedFile · function · L343-L351 — func typeBelongsToScannedFile(typeName *types.TypeName, files []sourceFile, fset *token.FileSet) bool
- implements · function · L353-L361 — func implements(t types.Type, iface *types.Interface) bool
- assertedInterface · function · L363-L375 — func assertedInterface(expr ast.Expr) string
- groupByPackage · function · L377-L396 — func groupByPackage(files []sourceFile) []packageGroup
- packageName · function · L398-L405 — func packageName(path string) (string, error)
- findGoFiles · function · L407-L430 — func findGoFiles(target string) ([]string, error)
- walkGoFiles · function · L432-L452 — func walkGoFiles(root string) ([]string, error)
- parseArgs · function · L454-L498 — func parseArgs(args []string) (options, error)
- parseNonNegativeInt · function · L500-L512 — func parseNonNegativeInt(s string) (int, error)
