---
version: 2024-11-20
---

# Expressing Issuer's Identity using WebPKI <!-- omit in toc -->

- [Generation of Signing Keys](#generation-of-signing-keys)
- [Creation of Subdomain](#creation-of-subdomain)
- [Requesting a TLS Certificate](#requesting-a-tls-certificate)
- [Representation of the Signing Key and the X.509 Certificate in JWT](#representation-of-the-signing-key-and-the-x509-certificate-in-jwt)
  - [`jwk` Header Parameter: Issuer's Signing Key and Certificate in the Protected Header](#jwk-header-parameter-issuers-signing-key-and-certificate-in-the-protected-header)
  - [`iss_jwk`: Issuer's Signing Key and Certificate](#iss_jwk-issuers-signing-key-and-certificate)
- [Verify the Issuer Identifier](#verify-the-issuer-identifier)
- [The role of Certificate Transparency](#the-role-of-certificate-transparency)
- [Related work](#related-work)

In OpenID Connect (OIDC) and OAuth, the value of the `iss` (issuer) JWT claim is used to identify the Authorization Server and to obtain information about the server's signing keys to verify the signature of an ID Token. Since there's a direct interaction between the OP and RP, the approach works well for short-lived credentials and is not subject to privacy risks. However, when credentials are presented to an RP using a digital wallet this approach may face limitations as issuers may rotate their keys, the keys may be unavailable due to limited internet connectivity.

We propose a model that binds a signing public key to a domain name using WebPKI.  WebPKI is one of the globally recognised identity frameworks. WebPKI enables the binding of a public key to a domain name as obtaining a WebPKI certificate requires proving ownership of both the domain name and the private key.

Services like [CertBot](https://certbot.eff.org/) make it easy to obtain WebPKI certificates. Additionally, the [Certificate Transparency](https://certificate.transparency.dev/) initiative ensures that certificates can be easily verified for existence, conformity, and validity.

## Generation of Signing Keys

Implementations MUST generate signing keys using one of the supported algorithms listed below. Signature algorithms are specified in [JWA (RFC 7518)](https://datatracker.ietf.org/doc/html/rfc7518). For the purposes of interoperability, Relying Parties MUST support algorithms marked as "Recommended". Specific use cases MAY define additional signing algorithms as necessary.

| Algorithm | Description                     | Status      |
| --------- | ------------------------------- | ----------- |
| RS256     | RSASSA-PKCS1-v1_5 using SHA-256 | Recommended |
| RS384     | RSASSA-PKCS1-v1_5 using SHA-384 | Optional    |
| RS512     | RSASSA-PKCS1-v1_5 using SHA-512 | Optional    |
| ES256     | ECDSA using P-256 and SHA-256   | Recommended |
| ES384     | ECDSA using P-384 and SHA-384   | Optional    |
| ES512     | ECDSA using P-521 and SHA-512   | Optional    |

## Creation of Subdomain

A subdomain MUST be created to allow certificate issuing services to verify control over both the domain name and the signing key.

- If issuer manages their own domain and signing keys, the subdomain MUST follow the format `jwt.iss.{issuer's domain name}`.  
  **Example**: `jwt.iss.example.com`
  
- If a third party (provider) manages issuer's domain name and signing keys, the subdomain MUST follow the format `jwt.iss-mt.{issuer's domain name}.{provider's domain name}`.  
  **Example**: `jwt.iss-mt.myproject.eu.example.com`

If issuer's service is managed, the issuer SHOULD publish a TXT record under their domain `authorize: jwt.iss-mt.{issuer's domain name}.{provider's domain name} authorization_date: {RFC 3339 UTC date and time}`. The entry is case insensitive and MUST NOT contain the `https` schema. Authorization date value MUST be the date and time of the authorization.

As the authorization is not verifiable, we should consider the following improvement proposal:

- Improvement proposal: signing key delegation using WebPKI
  - Signing key: provider-generated sk + issuer's signature of SHA(public key)+SHA(issuer-controlled pk)

## Requesting a TLS Certificate

The issuer (or provider) MUST obtain a WebPKI certificate from a recognized Certificate Authority (CA).

In the certificate:

- The **Common Name (CN)** field MUST be set to the fully qualified domain name (FQDN) of the issuer's domain following the subdomain schema outlined in the previous section.
- The **Subject Alternative Name (SAN)** DNS entry MUST also be set to the same FQDN as the CN.

This ensures that the certificate can be correctly linked to the issuer's identifier in the `iss` claim of the JWT.

## Representation of the Signing Key and the X.509 Certificate in JWT

Signing keys and certificates can be represented in a JWT protected header or in the JWT payload, in case the JWT header cannot be modified.

### `jwk` Header Parameter: Issuer's Signing Key and Certificate in the Protected Header

The `jwk` header parameter MUST be a JWK representing the signing public key. The parameter MUST contain the required JWK members of the signing key type, it MUST contain the JWK members defined below and it MAY contain other JWK members.

- `alg` (REQUIRED): Specifies the digital signature algorithm used, as defined in [JWA](https://datatracker.ietf.org/doc/html/rfc7518). The algorithm must be appropriate for the signing key's key type.
- `kty` (REQUIRED): Identifies the key type, as defined in [JWA](https://datatracker.ietf.org/doc/html/rfc7518). Common types include "EC" (Elliptic Curve) and "RSA".
- `use` (REQUIRED): MUST be set to `"sig"` to indicate that the key is intended for digital signatures, as defined in [JWS](https://datatracker.ietf.org/doc/html/rfc7517).
- `key_ops` (REQUIRED): Specifies the allowed operations for the key. MUST include `["verify"]` to indicate that the key is used for signature verification, as defined in [JWS](https://datatracker.ietf.org/doc/html/rfc7517).
- `x5c` (REQUIRED): Contains the full X.509 certificate chain, including the root certificate. The first certificate in the array MUST be the certificate bound to the signing key, as defined in [JWS](https://datatracker.ietf.org/doc/html/rfc7517).

Example:

```jsonc
{
  // JWT Protected header
  "jwk": {
    "alg": "ES256",  // Algorithm: Elliptic Curve with SHA-256
    "kty": "EC",     // Key Type: Elliptic Curve
    "use": "sig",    // Usage: Digital Signature
    "key_ops": ["verify"], // Key Operation: Signature Verification
    "crv": "P-256",  // Curve: P-256 (NIST standard curve)
    "x": "T4AdQSAmA14GZF3Ywg3jHLpHzU7RbRFE65p-cchJNyQ", // x-coordinate for the EC public key
    "y": "tN8GIeSeCbT2g2genGO1aqi-ajnZCJaKzJ2VVa5wRm0", // y-coordinate for the EC public key
    "x5c": ["MIIDQ...", "MIIDQ...", "MIIDQ..."] // X.509 Certificate Chain (Base64-encoded)
  }
}
```

### `iss_jwk`: Issuer's Signing Key and Certificate

Many authorization servers impose restrictions on modifying JWT header properties. To address this, we introduce a top-level JWT claim `iss_jwk`. This claim MUST follow the rules defined for the [`jwk` header parameter](#jwk-header-parameter-issuers-signing-key-and-certificate-in-the-protected-header).

Example:

```jsonc
{
  // JWT Payload
  "iss_jwk": {
    "alg": "ES256",  // Algorithm: Elliptic Curve with SHA-256
    "kty": "EC",     // Key Type: Elliptic Curve
    "use": "sig",    // Usage: Digital Signature
    "key_ops": ["verify"], // Key Operation: Signature Verification
    "crv": "P-256",  // Curve: P-256 (NIST standard curve)
    "x": "T4AdQSAmA14GZF3Ywg3jHLpHzU7RbRFE65p-cchJNyQ", // x-coordinate for the EC public key
    "y": "tN8GIeSeCbT2g2genGO1aqi-ajnZCJaKzJ2VVa5wRm0", // y-coordinate for the EC public key
    "x5c": ["MIIDQ...", "MIIDQ...", "MIIDQ..."] // X.509 Certificate Chain (Base64-encoded)
  }
}
```

## Verify the Issuer Identifier

To verify the link between the `iss` value and the signing public key, the verifier MUST perform the following steps

- Obtain the public key from the signing key certificate from the x5c JWK parameter.
- Validate the X.509 certificate chain. The root Certificate Authority MUST be one of the recognised WebPKI Certificate Authorities.
- Verify that the domain name in the `iss` value (without the `https://` schema) matches the `{issuer's domain name}` value in the Common Name (CN) and the Subject Alternative Name (SAN DNS) according to the schema:
  - `jwt.iss.{issuer's domain name}` for self-managed services
  - `jwt.iss-mt.{issuer's domain name}.{provider's domain name}` for managed services
- Verify that the certificate has not been revoked at the time of signature creation.

Issuer SHOULD add an authorization TXT record: `authorize: jwt.iss-mt.{issuer's domain name}.{provider's domain name}`. Note that with this we can only verify that the service provider is authorized at the verification time, but we don't have a guarantee that the authorization has been valid at the signature creation time.

## The role of Certificate Transparency

If IDT++ are long lived and we want to check whether the signing certificate was valid at the point of signing, we can rely on the [Certificate Transparency](https://certificate.transparency.dev/) and [crt.sh](https://crt.sh).

## Related work

- [Proof of Issuer Key Authority (PIKA)](https://www.ietf.org/archive/id/draft-barnes-oauth-pika-01.html)
