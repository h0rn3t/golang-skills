# evals/ab/_implement/_golden/catalog/golden_test.go

- goldenSource · struct · L13-L15 — goldenSource
- newGoldenSource · function · L17-L19 — func newGoldenSource() *goldenSource
- Get · method · L21-L34 — func (s *goldenSource) Get(sku string) (string, error)
- TestResolveErrorReachesEveryReason · function · L39-L63 — func TestResolveErrorReachesEveryReason(t *testing.T)
- TestResolveSkipsUnknownAndKeepsGoing · function · L65-L80 — func TestResolveSkipsUnknownAndKeepsGoing(t *testing.T)
- TestResolveCostsOneRoundTripPerSKU · function · L84-L99 — func TestResolveCostsOneRoundTripPerSKU(t *testing.T)
- TestResolveEmptyInput · function · L101-L109 — func TestResolveEmptyInput(t *testing.T)
