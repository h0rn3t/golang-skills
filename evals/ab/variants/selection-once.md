## Remove Duplication to the End

When several branches differ only in the constants they carry — a rate, a
percentage, a threshold — the duplication has two axes: the **selection**
(which case applies) and the **computation** (what is done with the value).
Separate them, then check three things before calling the step done:

1. Each literal appears once in the package.
2. Each selection over the same key appears once. Three functions that each
   switch on the same argument are one selection written three times; fold
   them into one lookup and let the functions read from it.
3. Each condition ladder appears once. Four copies of the same
   `if x >= a … if x >= b` ladder are one ladder with four pairs of values.

Removing one axis and leaving the other is the most common unfinished
refactor on this kind of code. Stop only when all three hold, or the remaining
copy has a reason in the report.

Pick the shape by the final code, call sites included: an exported accessor
that already performs the selection, before a new unexported helper; a
`switch` when the cases carry logic; a `map` or slice literal indexed by the
key when they carry only values. A table exists to delete the branches, not
to be serviced — no search helper, no method, no loop to rebuild a list that
was already a literal. If the lookup needs those, the `switch` was shorter.
Map iteration order is not source order, so an ordered literal stays a
literal. Error texts and the point where an unknown key fails do not move.
