# IDT++ - an ID Token profile with support for cryptographic binding, selective disclosure, and Web PKI

- Status: draft
- Version: 2024-12-05
- Previous version: N/A
- Proposed: 2024-09-01

## Introduction

ID Tokens (IDTs), defined as JSON Web Tokens (JWTs) containing claims about authentication events, are widely adopted through the OpenID Connect (OIDC) infrastructure. IDTs carry many important properties, such as built-in selective disclosure, option for repudiation and non-repudiation and they don't require revocation. However, since IDTs are typically audience-restricted and have short lifetimes, they are not well-suited for sharing with third parties, such as through digital wallets.

Verifiable credentials (VCs) are typically modelled as long-lived tokens that can be shared with third partied through digital wallets - similar to digitally signed documents, but with an ability to cryptographically prove ownership, selectively disclose information, and verify the identity and accreditations of an issuer. Examples of VCs are mobile driver's licence, digital diploma, digital student ID, and other credentials.

We are introducing an IDT profile: ID Token++ (IDT++), a profile designed for IDT so that they can be issued through OIDC Authorization Servers and having support for cryptographic binding to digital wallets, selective disclosure, and identification of issuers using WebPKI. With the IDT++ profile we can start issuing Verifiable Credentials using the existing OIDC infrastructure. Learn how IDT++ simplifies the issuance process building on technologies like Asynchronous Remote Key Generation (ARKG), SD-Cha-Cha, and WebPKI.

## Overview

IDT++ fits into the OpenID Connect (OIDC) protocol flow as depicted in the diagram below. Steps (1)-(5) are the [abstract OIDC protocol steps](https://openid.net/specs/openid-connect-core-1_0.html#Overview) and steps [a-d] are the additional steps that extend the OIDC flow.

```mermaid
sequenceDiagram
  actor eu as End-User
  participant wt as Wallet<br>RP
  participant rp as Relying Party<br>RP
  participant op as OpenID Provider<br>OP
  participant ip as IDT++ service<br>OP

  note over eu, op: User connects her wallet (a one-time process)
  eu -->> op: [a.1] Share Master Public Key
  op -->> ip: [a.2] Store Master Public Key
  note over eu, ip: IDT++ issuance
  rp ->> op: (1) AuthN Request
  op ->> eu: (2) AuthN & AuthZ
  op -->> ip: [b.1] Generate IDT++ Claims
  ip -->> op: [b.2] IDT++ Claims
  op ->> rp: (3) AuthN Response
  rp ->> op: (4) UserInfo Request
  op ->> rp: (5) UserInfo Response
  alt [c] IDT++ to Wallet
  rp -->> wt: Download IDT++
  else [d] Download and import to wallet
  rp -->> eu:  Receive IDT++
  eu -->> wt: Import IDT++
  end
```

**[a]** The End-User shares its master public key with the OpenID Provider (OP). This is a one-time process, similar to (dynamic) client registration. If the user already registered its wallet with the OP, this step can be skipped.  
**(1)** The RP (Client) sends a request to the OpenID Provider.  
**(2)** The OP authenticates the End-User and obtains authorization.  
**[b]** The OP extends the End-User claims set (ID Token) by adding IDT++ claims.  
**(3)** The OP responds with an IDT++ and usually an Access Token.  
**(4)** The RP can send a request with the Access Token to the UserInfo Endpoint.  
**(5)** The UserInfo Endpoint returns Claims about the End-User, including the IDT++ claims.  
**[c]** The user downloads the IDT++ to its wallet same as storing a file.  
**[d]** The user downloads the IDT++ as a file (or receives it via email or other communication channel) and imports it into its Wallet as a file.

Step [a] is a one-time process that occurs upon End-User's wallet registration with the OP. If End-User already registered its wallet with the OP, this step can be skipped. Step [b] is a step that's executed for every IDT++ creation. The IDT++ service can be an external service or integrated into the OP's Authorization Server and only generates and adds additional claims to the IDT. Steps [c-d] showcase different ways of how IDT++ can be stored in the wallet - directly, by fetching them from the RP (as storing files) or importing them manually upon receiving them via email or other communication channel.

## ID Token++

The main change that IDT++ makes to the OIDC to enable End-Users to use the IDTs with their wallet is the extension of the IDT data model with additional claims. IDT++ introduces four main capabilities:

- Enables users store their claims in their digital wallets and to present them to verifiers - see [Asynchronous proof-of-possession key derivation](./01_async-pop.md)
- Enables users to selectively disclose their claims - see [SD-Cha-Cha](./02_sd-cha-cha.md)
- Enables issuers to express their identity using WebPKI - see [Expressing Issuer's Identity using WebPKI](./03_iss-webpki.md)

The following claims extend the [OIDC ID Token data model](https://openid.net/specs/openid-connect-core-1_0.html#IDToken):

- cnf: [Confirmation Claim](https://www.rfc-editor.org/rfc/rfc7800.html#section-3.1)
  > By including a `cnf` (confirmation) claim in a JWT, the issuer of the
   JWT declares that the presenter possesses a particular key and that
   the recipient can cryptographically confirm that the presenter has
   possession of that key.  The value of the `cnf` claim is a JSON
   object and the members of that object identify the proof-of-
   possession key.
  >
  > MUST be the [JWK Confirmation Method](https://www.rfc-editor.org/rfc/rfc7800.html#section-3.2). The JWK MUST contain the required key members for a JWK of that key type, it MUST contain the `kid` member, and MAY contain other JWK members. The JWK MUST be the ARKG-derived public key.  
  >
  > REQUIRED for: Proof of Possession, Selective Disclosure SD-Cha-Cha
- kdk: Key Derivation Key
  > By including a `kdk` (key derivation key) claim in a JWT, the issuer of the JWT declares that it derived the presenter's proof-of-possession key using ARKG and that the `kdk` key MUST be used by the user to derive the corresponding private key.
  >
  > MUST be a JWK. The JWK MUST contain the required key members for a JWK of that key type.
  >
  > REQUIRED for: Proof of Possession, Selective Disclosure SD-Cha-Cha
- sdp: Selective Disclosure Parameters
  > `sdp` (selective disclosure parameters) claim in a JWT MUST be present if selective disclosure is used. The value of the claim is a JSON object that contains SD parameters and blinded claims.
  >
  > REQUIRED for: Selective Disclosure SD-Cha-Cha
- iss_id: Issuer identity
  > By including an `iss_id` claim in a JWT, the issuer of the JWT presents additional identity information about itself. WebPKI profile is defined in this document.
  >
  > REQUIRED for: WebPKI issuer ID.

Below we summarise which claims become REQUIRED, if a given capability is used:

| Capability           | Required claims |
| -------------------- | --------------- |
| Proof of Possession  | cnf, kdk        |
| Selective Disclosure | cnf, kdk, sdp   |
| WebPKI issuer ID     | iss_jwk         |

The capabilities can be used independently, hence use cases can decide which capabilities they use.

You can inspect an example [IDT++ on JWT.io](https://jwt.io/?id_token=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IjVSNXZlZ2cyMjMycDIweC1hM3c4USJ9.eyJodHRwOi8vY29uc3VsLmludGVybmFsL2ZpcnN0X25hbWUiOiIiLCJodHRwOi8vY29uc3VsLmludGVybmFsL2xhc3RfbmFtZSI6IiIsImh0dHA6Ly9jb25zdWwuaW50ZXJuYWwvZ3JvdXBzIjpbXSwibXluZXh0LmlkLmNyZWRlbnRpYWwiOnsiY25mIjp7ImNydiI6IlAtMjU2Iiwia3R5IjoiRUMiLCJ4IjoibUZic0Q4NWcwS2hRN19CYV9ESXVQcldsNGRmVllRNjdOS2w4U3hPVUxncyIsInkiOiJGNzJ0dlVPU0RvYnptdjVZT2NKYXo1eHJVLWdQQmhyaDRHN18zSERfVUE4In0sImNyZWRlbnRpYWxTdWJqZWN0Ijp7ImlkIjoiMTIzIiwiaXNzdWVyIjoiTmV0aXMgZC5vLm8uIiwibWFzdGVyUGFzc3dvcmQiOiIiLCJudWNsZWFyQ29kZXMiOltdLCJwcm9qZWN0SWQiOiJJRFQrK0RFTU8ifSwia2giOnsiY3J2IjoiUC0yNTYiLCJrdHkiOiJFQyIsIngiOiJoQzFpQk5WYVRhTWlJcE1KMHFpZktPZHRneHFkSFEyMG1BbEdLLXdPU0tFIiwieSI6IjJsaXFVVXR3NVV1MVlVWF9sWjRhYUNWaEVOSUFzQXY4X1MyTVJ6YkNtcWcifSwic2RwIjp7ImFsZyI6InNkLWNoYS1jaGEiLCJhcHUiOiIiLCJhcHYiOiIiLCJiYyI6eyIvY3JlZGVudGlhbFN1YmplY3QvbWFzdGVyUGFzc3dvcmQiOiJwdkk3eVpzemduVHVZUDI4N3dfRHdmNDJ0UmtlZUV4WjVjajVzeVNkIiwiL2NyZWRlbnRpYWxTdWJqZWN0L251Y2xlYXJDb2RlcyI6IkpJN2dDOTBLUllKcFh3eDVMVXZYNnlsSUtMdkpPb2p1S09yQ2phVllEVUZCIn0sImVuYyI6IlhDaGFDaGEyMCIsImVwayI6eyJjcnYiOiJQLTI1NiIsImt0eSI6IkVDIiwieCI6IjZGa1BsbXhiWTlLczJ1RGRMZmxzOHdGLWVCV0hreFF1NF81OGx5cjNoZ2siLCJ5IjoiUXprSVZFQTRUelRUQkxEdlZYa2ZKbDVPbVpQdTM1SkpVQ3FBTDU4azNGYyJ9LCJoYXNoIjoiU0hBLTI1NiJ9fSwiZ2l2ZW5fbmFtZSI6IkFsZW4iLCJmYW1pbHlfbmFtZSI6IkhvcnZhdCIsIm5pY2tuYW1lIjoiYWxlbi5ob3J2YXQubmV0aXMiLCJuYW1lIjoiQWxlbiBIb3J2YXQiLCJwaWN0dXJlIjoiaHR0cHM6Ly9saDMuZ29vZ2xldXNlcmNvbnRlbnQuY29tL2EvQUNnOG9jSVBNMXNyRnpoOTlxSUh6TWhUcWJoWmZWd0hieGNxMENGeHlwbWoxN19JR1NsVndrTG49czk2LWMiLCJ1cGRhdGVkX2F0IjoiMjAyNC0xMS0yMFQxNDowODo1NS40NzJaIiwiaXNzIjoiaHR0cHM6Ly9kZXYtazcxMjNoeXQudXMuYXV0aDAuY29tLyIsImF1ZCI6Ikk5dDZPU1IzbWRFZnBycnNtT0FERFVXYTlHM2FWb1F3IiwiaWF0IjoxNzMyMTExNzM4LCJleHAiOjE3MzIxNDc3MzgsInN1YiI6Imdvb2dsZS1vYXV0aDJ8MTAyMTE3OTkyNTMxMTE1ODIxNTE5Iiwic2lkIjoiNkgtbEd3cjUyVTJJZXZtSnFFVW40OUsyNmZWcGdLY0cifQ.lXpB7AILw6nM7Daq7aOCqQlKJ2OO_dMpKD50jjQ-WaIR2KlFtZjVibfibCiGvFyPDudTsshoGakzaloejJXlGCcUtotO7Z39rOWGEyDakGFCFX8z5Zju1PwylSfMPqu4wI3yPd6cIy6BzjdD9gViNTB16Bk_0OqqQfZV187iu-1DDuSOCKdsHl5s1d9RwNrjoRzsSJd_ktJOgDolkfnE3ye_UsZlU9WUmTG6OYRxidKKNwv3HxwMnPecXkzNGtRhcXDqylXuNOa7kDtDN4gMoY9ScFuCKbp9PrMYRqSJxcDCAfWgXVuy5OVFSFjKWlZyqwgYFWmvhJE11WP96GAuNw)
or you can obtain your IDT++ using the [IDT++ Demo App](https://test-api.mynext.id/idt/v2/).

## Specifications

In the documents below you can find details about the different capabilities.

- [Asynchronous Proof-of-Possession Key Generation](./01_async-pop.md)
- [Selective Disclosure Cha-Cha](./02_sd-cha-cha.md)
- [Issuer identity with WebPKI](./03_sd-cha-cha.md)
- [Appendix](./99_appendix.md)

## Demo

- [MyNextID IDT++ showcase](https://test-api.mynext.id/idt/v2/)

## Reference implementation

- Coming soon.

## Extensions

- IDT++ file server: Once the wallet is connected, an additional IDT++ OAuth-protected (e.g., private_key_jwt) file-server can be introduced where the wallet can fetch "fresh" credentials as files (in different formats and signatures).
- Extension for PQC algorithm
- Extension for BSS+ or other selective disclosures

## Considerations

The design is not limited to JWT signature profile and can be applied to other credential and signature formats, such as mdoc, W3C Verifiable Credentials, JAdES, and others.

## License

See the [LICENSE](./LICENSE)

## Acknowledgements

This work is being supported by the world greatest team at Netis.
