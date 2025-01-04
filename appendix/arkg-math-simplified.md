# Asynchronous Remote Key Generation (ARKG) - Math Simplified

ARKG has a simple yet powerful property: Alice can create Bob's public keys
without knowing Bob's private key (which corresponds to the generated public
keys). Furthermore, all the information can be exchange securely via any public
communciation channel. This capability is particularly useful for applications
such as account recovery and credential issuance.

In this document, you will learn about the mathematics behind ARKG using
Elliptic Curves. See [Elliptic Curve Cryptography 101](ec-101.md) for an
introduction to the topic.

## ARKG Simplified

The process can be outlined as follows:

### Step 1: Bob Creates a Master Key Pair

Bob begins by creating a master public-secret key pair:

```pseudocode
(mpk, msk) = generateECKeyPair()
```

He then shares his master public key `mpk` with Alice.

### Step 2: Alice Creates an Ephemeral Key Pair

To generate a new public key for Bob, Alice creates an ephemeral public-secret key pair:

```pseudocode
(epk, esk) = generateECKeyPair()
```

### Step 3: Alice Computes a Shared Secret

Next, Alice computes a shared secret using the Diffie-Hellman key exchange:

```pseudocode
sharedSecret = ECDH(esk, mpk) = esk * mpk mod N
```

Bob can compute the same shared secret using:

```pseudocode
sharedSecret = ECDH(msk, epk) = ECDH(esk, mpk)
```

This is why Alice will need to share with the Bob the `epk`.

### Step 4: Alice Derives a Public Key

Using the shared secret as the secret key, Alice derives a new public key:

```pseudocode
pk_ss = sharedSecret * G mod N
```

She then calculates the derived public key by combining the master public key and the shared secret public key:

```pseudocode
pk_derived = mpk + pk_ss
```

Alice shares the ephemeral public key (`epk`) and the derived public key
(`pk_derived`) with Bob. As Alice is sharing only public information with Bob,
the public keys can be shared securely via any public communication channel.

### Step 5: Bob Computes the Shared Secret and Derived Private Key

Using the provided `epk`, Bob computes the shared secret:

```pseudocode
sharedsecret = epk * msk
```

Then, he derives the private key:

```pseudocode
sk_derived = msk + sharedSecret
```

### Step 6: Verification of the Public Key

Finally, the correctness of the public key can be verified:

```pseudocode
pk_derived' = sk_derived * G
             = (msk + sharedSecret) * G
             = mpk + sharedSecret * G
             = mpk + pk_ss
             = pk_derived
```

Note: This explanation provides a simplified summary of the math behind ARKG. For more details, please refer to the references.

## Potential use cases

- Account recovery
- Asynchronous Verifiable Credential issuance
- Messaging
  - Open question: can this be used to replace Signal's central key distribution server?
- Blockchain transactions
- other

## References

- <https://www.yubico.com/blog/yubico-proposes-webauthn-protocol-extension-to-simplify-backup-security-keys/>
- <https://hackernoon.com/blockchain-privacy-enhancing-technology-series-stealth-address-i-c8a3eb4e4e43>
- <https://github.com/w3c/webauthn/issues/1640>
- <https://github.com/Yubico/webauthn-recovery-extension>
- <https://www.ietf.org/archive/id/draft-bradleylundberg-cfrg-arkg-02.html>
- <https://github.com/Yubico/webauthn-recovery-extension/tree/master/benchmarks>
