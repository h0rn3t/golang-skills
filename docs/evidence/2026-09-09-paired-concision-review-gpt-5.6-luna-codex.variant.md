## Deletion Review

Before finishing, apply the deletion test to each unexported helper, type, and
layer in the result. Inline it when doing so makes its complexity disappear.
Keep it when inlining would spread a shared rule or meaningful operation across
callers. Finish with only abstractions justified by current variation or
callers; remove pass-through middle men and speculative generality. Judge the
complete code and call sites, and stop when another deletion would reduce
clarity or change behavior.
