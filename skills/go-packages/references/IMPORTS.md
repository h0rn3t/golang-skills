# Import Organization

> Sources: source/golang-wiki/CodeReviewComments.md (Imports, Import Blank, Import Dot); source/google-go-styleguide/decisions.md (Import "blank"); source/uber-go-style/style.md (Import Group Ordering)
> Authority: advisory
> Last verified: 2026-09-10

## Import Grouping

Imports are organized in groups, with blank lines between them. The standard
library packages are always in the first group.

**Minimal grouping (Uber):** stdlib, then everything else.

**Extended grouping (Google):** stdlib → other → protocol buffers → side-effects.

```go
// Good: Full grouping with protos and side-effects
import (
    "fmt"
    "os"

    "github.com/dsnet/compress/flate"
    "golang.org/x/text/encoding"

    foopb "myproj/foo/proto/proto"

    _ "myproj/rpc/protocols/dial"
)
```

## Import Renaming

Avoid renaming imports except to avoid a name collision; good package names
should not require renaming. In the event of collision, **prefer to rename the
most local or project-specific import**.

**Must rename:** collision with other imports, generated protocol buffer packages
(remove underscores, add `pb` suffix).

**May rename:** uninformative names (e.g., `v1`), collision with local variable.

## Blank Imports (`import _`)

Packages that are imported only for their side effects (using `import _ "pkg"`)
should only be imported in the main package of a program, or in tests that
require them.

```go
// Good: Blank import in main package
package main

import (
    _ "time/tzdata"
    _ "image/jpeg"
)
```

The one blank import a library file carries is `embed`, in a file whose
`//go:embed` fills a `string` or `[]byte`; without it the build fails with
`go:embed requires import "embed"`:

```go
// Good: a library file embedding into a string
package schema

import _ "embed"

//go:embed schema.sql
var DDL string
```

## Dot Imports (`import .`)

**Do not** use dot imports. They make programs much harder to read because it is
unclear whether a name like `Quux` is a top-level identifier in the current
package or in an imported package.

**Exception:** The `import .` form can be useful in tests that, due to circular
dependencies, cannot be made part of the package being tested:

```go
package foo_test

import (
    "bar/testutil" // also imports "foo"
    . "foo"
)
```
