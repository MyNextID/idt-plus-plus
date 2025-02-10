# Dynamic Status List for JWTs

Revocation means trouble. Don't use it if you don't have to. If you have to
define a revocation strategy, here we're presenting a simple and
easy-to-implement revocation profile for the Dynamic Status List approach as
defined in EBSI (European Blockchain Services Infrastructure).

Note: this profile and implementation are for education purposes only and
parameters are not optimized for production.

Note: this is an early draft. Open an issue to report bugs, questions, proposals, etc.

## What are the key properties of the DSL for JWT?

- time-dependent status visibility
- embedded or detached status list information
- encrypted revocation metadata (v2)
  - exact revocation time
  - revocation reason
  - other extensions
- binding enhancement using ARKG

## How it works?

See the [specifications](dsl-jwt.md)

## Quick start

To get your hands dirty, check out the simple [CLI](cli) or run the [script](cli/script.sh)

## Compatibility

Note that the design is not limited to JWT format and data model.

## Acknowledgements

This work is a JWT profile for the Dynamic Status List introduced in [EBSI](https://hub.ebsi.eu/vc-framework/credential-status-framework).

## Appendix - Helper functions

### Pretty-print JSON

To pretty print a JSON, run

```bash
./dsl print -i {path to a JSON file}
```

### Pretty-print JWT

To pretty print a JWT, run

```bash
./dsl printjwt -i '{path to a (JSON) file containing JWT as utf-8 encoded string or JSON object with a jwt claim}'
```
