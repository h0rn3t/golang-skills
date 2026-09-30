# Receiver Type: Pointer vs Value

> Sources: source/golang-wiki/CodeReviewComments.md (Receiver Type)
> Authority: advisory
> Last verified: 2026-09-10

Choosing whether to use a value or pointer receiver on methods can be difficult.
**If in doubt, use a pointer**, but there are times when a value receiver makes
sense.

## When to Use Pointer Receiver

- **Method mutates receiver**: The receiver must be a pointer
- **Receiver contains sync.Mutex or similar**: Must be a pointer to avoid copying
- **Large struct or array**: A pointer receiver is more efficient. If passing all
  elements as arguments feels too large, it's too large for a value receiver
- **Concurrent or called methods might mutate**: If changes must be visible in
  the original receiver, it must be a pointer
- **Elements are pointers to something mutating**: Prefer pointer receiver to
  make the intention clearer

## When to Use Value Receiver

- **Small unchanging structs or basic types**: Value receiver for efficiency
- **Map, func, or chan**: Don't use a pointer to them
- **Slice without reslicing/reallocating**: Don't use a pointer if the method
  doesn't reslice or reallocate the slice
- **Small value types with no mutable fields**: Types like `time.Time` with no
  mutable fields and no pointers work well as value receivers
- **Simple basic types**: `int`, `string`, etc.

## Consistency Rule

**Don't mix receiver types**. Choose either pointers or struct types for all
available methods on a type. If any method needs a pointer receiver, use pointer
receivers for all methods.
