# Appendix <!-- omit in toc -->

- [Appendix A. References](#appendix-a-references)
  - [Normative references](#normative-references)
  - [Informative references](#informative-references)
- [Appendix B. Implementations](#appendix-b-implementations)
- [Appendix C. Document History](#appendix-c-document-history)
- [Appendix D. Terminology](#appendix-d-terminology)
  - [Actors](#actors)

> [!NOTE]
> This document has not been edited nor finalized.

## Appendix A. References

### Normative references

- [ARKG](https://www.ietf.org/archive/id/draft-bradleylundberg-cfrg-arkg-02.html) - The Asynchronous Remote Key Generation (ARKG) algorithm - Draft 02
- [OIDC](https://openid.net/specs/openid-connect-core-1_0.html) - OpenID Connect Core 1.0
- [RFC 7517](https://datatracker.ietf.org/doc/html/rfc7517) - JSON Web Key (JWK)
- [RFC 7518](https://datatracker.ietf.org/doc/html/rfc7518) - JSON Web Algorithms (JWA)

### Informative references

- [OIDC UserInfo VCs](https://github.com/bifurcation/userinfo-vc/blob/main/openid-connect-userinfo-vc-1_0.md) - OpenID Connect UserInfo Verifiable Credentials - Draft 00 proposed a profile for providing userinfo in a Verifiable Credentials profile.

## Appendix B. Implementations

- [MyNextID Showcase](https://test-api.mynext.id/idt)

## Appendix C. Document History

- 2024-11-20: Major refactoring (v1 - core capabilities, v2 - extensions)
- 2024-09-08: Initial version

## Appendix D. Terminology

> [!NOTE] Work In progress

This specification uses terms Relying Party (RP), OpenID Provider (OP), Authorization Server (AS), JSON Web Token (JWT), ID Token (IDT), JSON Web Signature (JWS), as defined in [OpenID Connect](https://openid.net/specs/openid-connect-core-1_0.html).

This specification uses the following terms:

- **Asynchronous Remote Key Generation (ARKG)**: ARKG Definition

- **Authentic Source (issuer)**: An entity capable of making claims about a resource owner and providing the information to the resource server.  

- **Authorization Server**: The server issuing access tokens to the client after successfully authenticating the resource owner and obtaining authorization. ref_RFC6749

- **Client (Wallet)**: An application making protected resource requests on behalf of the resource owner and with its authorization. ref_RFC6749

- **Resource Owner (Holder)**: An entity capable of granting access to a protected resource. ref_RFC6749

- **Resource Server**: The server hosting the protected resources, capable of accepting and responding to protected resource requests using access tokens. ref_RFC6749

- **Verifiable Credential**: A digitally signed object or data structure that authoritatively binds an identity and/or additional attributes of a subject to a digital wallet controlled by a holder.

- **Verifier**: Entity requesting, receiving and verifying Verifiable Credentials.

- **Wallet**: A service or application under user's control for requesting and presenting Verifiable Credentials.

- **Key Encapsulation Mechanism (KEM)**: is a public-key cryptosystem that allows a sender to generate a short secret key and transmit it to a receiver securely, in spite of eavesdropping and intercepting adversaries.  

- **Key Blinding (BL)**: is the process by which a private signing key or public verification key is blinded (randomized) to hide information about the key pair.

- **Elliptic Curve Diffie-Hellman (ECDH)**: Define ECDH here.

### Actors

RFC7800:

**Issuer**
: Entity that creates the JWT and binds the proof-of-possession key to it.  
**Presenter**
: Entity that proves possession of a private key (for asymmetric key cryptography) or secret key (for symmetric key cryptography) to a recipient.  
**Recipient**
: Entity that receives the JWT containing the proof-of-possession key information from the presenter.

Wallet

- Issuer
- Holder
- Verifier

OIDC

- OpenID Provider (OP): OAuth 2.0 Authorization Server that is capable of Authenticating the End-User and providing Claims to a Relying Party about the Authentication event and the End-User.
- Relying Party: OAuth 2.0 Client application requiring End-User Authentication and Claims from an OpenID Provider.
- Claims Provider - Server that can return Claims about an Entity.

- Claim: Piece of information asserted about an Entity.
- ID Token - JSON Web Token (JWT) [JWT] that contains Claims about the Authentication event. It MAY contain other Claims.
- Issuer: Entity that issues a set of Claims.

OAuth2

**Client**
: An application making protected resource requests on behalf of the resource owner and with its authorization.  The term "client" does not imply any particular implementation characteristics (e.g., whether the application executes on a server, a desktop, or other devices).  
**Resource Server**
: Server hosting the protected resources, capable of accepting and responding to protected resource requests using access tokens.  
**Authorization Server**
: The server issuing access tokens to the client after successfully authenticating the resource owner and obtaining authorization.  
**Resource Owner**
: An entity capable of granting access to a protected resource.  When the resource owner is a person, it is referred to as an end-user.

Component view:

- OIDC Authorization Server
- Wallet
- Relying Party
