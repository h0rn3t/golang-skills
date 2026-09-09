# skills/go-error-handling/scripts/check-errors-ast.go

- finding · struct · L18-L23 — finding
- options · struct · L25-L32 — options
- usage · function · L34-L47 — func usage()
- main · function · L49-L137 — func main()
- analyzeFile · function · L139-L223 — func analyzeFile(path string, checkBareReturn bool) ([]finding, error)
- isErrorCall · function · L225-L232 — func isErrorCall(expr ast.Expr) bool
- isStringLiteral · function · L234-L237 — func isStringLiteral(expr ast.Expr) bool
- isStringsContainsErrorCall · function · L239-L249 — func isStringsContainsErrorCall(call *ast.CallExpr) bool
- containsErrorCall · function · L251-L264 — func containsErrorCall(expr ast.Expr) bool
- isLogCallWithErr · function · L266-L286 — func isLogCallWithErr(call *ast.CallExpr) bool
- containsIdent · function · L288-L302 — func containsIdent(expr ast.Expr, name string) bool
- returnsErr · function · L304-L310 — func returnsErr(stmt *ast.ReturnStmt) bool
- findGoFiles · function · L312-L335 — func findGoFiles(target string) ([]string, error)
- walkGoFiles · function · L337-L357 — func walkGoFiles(root string) ([]string, error)
- parseArgs · function · L359-L403 — func parseArgs(args []string) (options, error)
- parseNonNegativeInt · function · L405-L417 — func parseNonNegativeInt(s string) (int, error)
