## Concision Gate

This pass succeeds only when the final production code reads at least as
clearly as the starting code and contains no more production lines. Measure the
starting and final production line counts. Keep a transformation only when both
conditions hold. If none does, restore the starting code: an empty diff is a
successful concision result. Preserve validation, failure behavior, security
controls, and useful abstraction boundaries, and report the two counts.
