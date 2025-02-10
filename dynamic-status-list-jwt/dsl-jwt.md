# Dynamic Status List for JWT

A JWT-based solution for managing a Dynamic Status List (DSL).

## Why It Works

A Status List is essentially a collection of identifiers or serial numbers
associated with issued and/or revoked credentials. A Dynamic Status List (DSL)
operates differently in that the identifiers are recomputed periodically (every
`dt` period). If you know the status identifier at a specific time `t`, you will
not find it in the list at a later time `t + dt`, as the issuer will have
recomputed the identifiers. Furthermore, you won't be able to predict the new
identifier unless there is only one identifier in the list. Essentially, these
identifiers are time-based tokens, which prevent verifiers from tracking status
changes unless explicitly granted by the holder.

The sequence of these time-based tokens can only be computed by the issuer and
the credential holder, both of whom share the same secret `seed`. Each
credential must have its own unique secret.

## How It Works

When the issuer creates a credential, they generate a random secret `seed`,
which is used to compute the time-based token. The token is calculated as
follows:

```javascript
t = Floor(timeNow / dt)
token = HMAC(seed, t)
```

The status list entry for a valid credential is then computed as:

```javascript
sid = SHA256(token, jti)
```

Where `jti` refers to the [JWT Identifier](https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.7).

If the credential is revoked, the entry is computed as:

```javascript
sid = SHA256(SHA256(token, jti))
```

The user shares both the `token` and `jti` with the verifier so that the
verifier can compute the corresponding `sid` and validate the credential.

If the `sid` is not found in the status list, the verifier MUST reject the
credential as the holder is presenting an invalid token.

## Limitations

- To check the status at a past time, the issuer must either retain or recompute
the historical status information on demand.
- If the seed is compromised, the JWT status can be recalculated and potentially
tampered with.

## Extension: Encrypted Status Metadata

The status metadata can be encrypted for additional privacy and security,
ensuring that only authorized parties can access the details associated with the
credential status.

## Advanced: Enhancing Security with Shared Secrets and ARKG

Idea:

- By leveraging the user’s public key, it’s possible to create an identifier
such that:
  - Anyone who knows the public key can compute the identifier if it’s in the
  form `ARKG(pk, t)`, but this is not ideal for privacy.
  - A better approach involves using both the public key and the seed. This
  creates a stronger binding between the identifier and the holder’s key.
    - Optimization: The seed can be a shared secret between the holder and
    the issuer, assuming both parties use compatible cryptographic systems.
