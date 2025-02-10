# Dynamic Status List for JWT

A JWT profile for the Dynamic Status List.

## How it works?

WIP

## Why it works?

WIP

## What are the limitations?

- Checking the status in the past requires the issuer to hold or on-demand recompute the historical status information.
- If the seed is revealed, JWT status can be re-computed

## Extension: encrypted status metadata

## Advanced: improving the security with shared secrets and ARKG

Idea:

- Using user's public key, we can always
  - everyone who knows the public key can compute the identifier if the identifier is ARKG(pk, t): this won't work
  - public key + seed -> helps as we bind the check to the holder's binding key
    - optimization: seed can be the shared secret between the holder and the issuer (assuming both use the same cryptography)
