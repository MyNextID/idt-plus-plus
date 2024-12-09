# Extension: IDT++ Server

> [!NOTE] This document outlines the main idea.

Once the wallet is connected, an additional IDT++ OAuth-protected (e.g., private_key_jwt) file-server can be introduced where the wallet can fetch "fresh" credentials as files (in different formats and signatures).

Master Public Key registered with the issuer should be used for private_key_jwt authentication.
