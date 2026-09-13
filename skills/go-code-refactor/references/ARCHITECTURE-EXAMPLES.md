# Architecture Examples from Community Repositories

> Sources: [vahiiiid/go-rest-api-boilerplate](https://github.com/vahiiiid/go-rest-api-boilerplate); [garvishtayal/go-gin-gorm-starter](https://github.com/garvishtayal/go-gin-gorm-starter); [sarrooo/go-clean](https://github.com/sarrooo/go-clean); [bxcodec/go-clean-arch](https://github.com/bxcodec/go-clean-arch) — READMEs read on 2026-09-13
> Authority: advisory — evidence that the vocabulary in ARCHITECTURE.md is in use, not templates to copy or production-quality certifications
> Last verified: 2026-09-13; inspect a specific revision before relying on a repository's current structure

Companion to [ARCHITECTURE.md](ARCHITECTURE.md). Read it when an example is
needed to explain a choice, not before every refactor. Each repository is
described as its README described it on the date above; structures change.


These examples are evidence that the naming is workable, not authorities that
must be copied verbatim.

### Feature-first + Handler/Service/Repository: GRAB

[vahiiiid/go-rest-api-boilerplate](https://github.com/vahiiiid/go-rest-api-boilerplate)
explicitly documents both:

- `Handler -> Service -> Repository` as the request flow;
- domain-driven organization by feature rather than one global layer tree.

Its `internal/user/` domain keeps `model.go`, `repository.go`, `service.go`, and
`handler.go` together. For a larger domain, turning those files into
`models/`, `repositories/`, `services/`, and `handlers/` subpackages is a natural
next split without changing the ownership model.

### Classic global layers: go-gin-gorm-starter

[garvishtayal/go-gin-gorm-starter](https://github.com/garvishtayal/go-gin-gorm-starter)
uses the familiar top-level shape:

```text
internal/
    handlers/
    services/
    repositories/
    domain/
```

This is easy to navigate for a small, cohesive service. Its scaling limit is
not the layer names; it is when many unrelated domains begin sharing each layer.

### Classic `models/services/repositories` with GORM: go-clean

[sarrooo/go-clean](https://github.com/sarrooo/go-clean) uses controllers,
services, repositories, and models with GORM. It demonstrates the common
layered vocabulary and the repository abstraction. Its global layout is most
appropriate while the application remains cohesive enough that the layer
packages do not become domain grab-bags.

### Consumer-side interfaces: bxcodec/go-clean-arch

[bxcodec/go-clean-arch](https://github.com/bxcodec/go-clean-arch) is useful for a
different reason: its newer structure explicitly moved interfaces toward the
consuming side and toward service-focused packages. That supports the rule used
here: keep `repositories/` as concrete data access while a service declares only
the small storage behavior it actually needs.

The lesson from these projects is not that one directory spelling is "clean
architecture". The repeatable properties are:

- thin handlers;
- services own business/use-case orchestration;
- repositories own data access;
- dependency injection at startup;
- interfaces at real seams;
- feature/domain ownership before a large codebase turns global layer packages
  into dumping grounds.
