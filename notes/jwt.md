# JWT (JSON Web Tokens)

## What a JWT is

A JWT is a defined standard (RFC 7519) for securing information exchange. In an auth flow: you log in, the server returns a token, and each subsequent request carries that token to demonstrate and prove who is talking to the server, until the token expires.

## The three components

A JWT is three base64url-encoded parts joined by dots: `header.payload.signature`.

- **Header**: information about the type and the algorithm used to produce the signature.
- **Payload**: contains the **claims**. Each registered claim is one statement. If the subject is "John Doe", the statement is "this token is about John Doe".
- **Signature**: created from all the information before it. You take the algorithm and pass it the header, the payload, and the secret key.

## The critical misconception: the payload is NOT encrypted

Anyone holding the token can base64-decode the payload and read every claim, **with no key at all**. Paste any token into jwt.io and the claims are in plaintext.

So the secret does **not** protect confidentiality and does not "open" anything. It protects:

- **Authenticity**: this token was issued by someone holding the secret, so by your server.
- **Integrity**: the claims have not been altered since issuance.

The right framing: **it is not a security layer in the sense of "hide a secret", it is a security layer in the sense of "verify who sent you this information".** A JWT is a **signed document**, not a sealed envelope. Anyone can read it; only the holder of the secret can produce a valid signature.

Consequence: never put anything secret in claims. No passwords, no keys, nothing you would not hand to the client, because you are handing it to the client.

(When an encrypted payload is genuinely needed there is a separate standard, JWE, but it is rarely what web auth uses.)

## How verification actually works

The signature is `HMAC-SHA256(header + "." + payload, secret)`. HMAC is deterministic: same input, same secret, same output, always.

**The signature travels inside the token**, as its third part. The server does not store it or look it up anywhere.

When a token arrives, the server:

1. Splits it on the dots into header, payload, signature.
2. Takes the header and payload **as received** and recomputes `HMAC-SHA256(header + "." + payload, secret)` with its own copy of the secret.
3. Compares its result against the signature that arrived.

Match means the token was signed with that secret and nothing was altered. Mismatch means it was not signed by you, or a byte was changed.

The secret is never transmitted. It exists only on the server, on both the signing side and the verifying side. It is an ingredient in a computation, not a decryption key.

## The attacker's perspective

An attacker controls all three parts of what they send. They can rewrite the payload and rewrite the signature freely. The reason forgery fails is not that they cannot touch the signature, it is that they cannot compute the **correct** one without the secret, and the server's recomputation catches the mismatch.

A different payload produces a different signature. So modifying `sub` to another user's ID requires the signature corresponding to *that* modified payload. Their options, all dead ends:

- Keep the original signature with a modified payload → recomputation differs, rejected.
- Compute a new signature for the modified payload → needs the secret. Blocked.
- Guess the signature → 256 bits, not feasible.
- Derive the secret from a token they hold → HMAC-SHA256 is not invertible. Blocked.

**With the secret, an attacker can forge whatever they like and it verifies perfectly.** The secret is the entire security boundary: environment variable, never in the repo. Rotating it invalidates every token ever issued.

### What an attacker can actually do

The real attack surfaces bypass the signature rather than break it:

- **Steal a valid token** (XSS, a leaky log, an insecure channel) and replay it as-is. It is genuinely valid until `exp`. Hence short lifetimes, HTTPS, HttpOnly cookies.
- **Attack the `alg` field**: set it to `none` and send no signature, hoping the library skips verification; or swap RS256 to HS256 so the server verifies using the public key as an HMAC secret, which the attacker knows. Both are library bugs, not crypto weaknesses.
- **Brute-force a weak secret**: offline cracking against a captured token is fast if the secret is short. Use 32+ bytes from a CSPRNG.

## Claims

A claim is one statement inside the payload. `jwt.RegisteredClaims` holds the standardized set from RFC 7519: `iss`, `sub`, `aud`, `exp`, `nbf`, `iat`, `jti`. Standardized means any JWT library in any language validates them the same way. Custom claims go in your own struct embedding `RegisteredClaims`.

## Signing methods

The algorithm used to produce the signature. The important axis is symmetric vs asymmetric:

- **HS256** (HMAC + SHA-256): **symmetric**, the same secret signs and verifies. Correct when one service does both.
- **RS256** (RSA): **asymmetric**, a private key signs and a public key verifies. Needed when several parties must verify but only one may issue.

## Creating a token

```go
func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	})

	return token.SignedString([]byte(tokenSecret))
}
```

`SignedString` does everything at once: base64url-encodes the header and payload, joins them with a dot, HMACs that string with the secret, and appends the encoded signature. There is no separate "assemble components, then sign" step to write.

Times must be wrapped in `jwt.NewNumericDate`, not passed as raw `time.Time`.

## Validating a token: pin the algorithm

Always specify which signing methods are acceptable, to defeat the `alg` attacks above:

```go
jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, keyFunc,
	jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
```

Validation means both recomputing the signature and checking the claims (`exp` in the past, `iss` as expected). The library performs the expiry check during parsing, which is why an expired token comes back as an error rather than valid-but-stale.

## The request flow

1. User logs in with credentials.
2. Server verifies them and returns a signed JWT.
3. Client sends it on every subsequent request in a header: `Authorization: Bearer <token>`
4. Server reads `r.Header.Get("Authorization")`, strips the `Bearer ` prefix, validates the rest.
5. Repeat until the token expires, then the client gets a 401.

## Where the user ID comes from (a mistake worth recording)

**Once authentication is by token, the client no longer sends a user ID in the request body, and the server must never trust one if it does.**

The authenticated identity comes from the **`sub` claim of the validated token**, and only from there. A user ID in the body is attacker-controlled: anyone could put someone else's ID in it. The `sub` claim is signature-protected, so it cannot be altered without the secret.

Rule: after validation, read the user ID out of the claims, and ignore any ID the request body offers.

## Where validation lives

Not copy-pasted into every protected handler. It belongs in **middleware**, a function taking a `Handler` and returning a `Handler` that validates first and either rejects with 401 or delegates. Same shape as `http.StripPrefix`.

The wrinkle: once the middleware has extracted the user ID, the handler needs it. Passing it down the chain is what `context` is for.

## The stateless tradeoff

With classic sessions, the server stores a session ID and looks it up on every request. With a JWT the server stores **nothing**: everything needed is in the token, and the signature proves it is trustworthy. Verification is one HMAC computation, no database hit.

The cost: **a JWT cannot be revoked.** There is no server-side record to delete, so a stolen token stays valid until `exp`.

That is the unresolvable tension with a single token: a long expiry means a stolen token is usable for a long time, a short expiry means the user is logged out constantly. The resolution is a **short-lived access token** (minutes) paired with a **long-lived refresh token** that *is* stored server-side and therefore *can* be revoked.

## References

- Introduction and when to use JWTs: https://www.jwt.io/introduction#when-to-use-json-web-tokens
- API: https://pkg.go.dev/github.com/golang-jwt/jwt/v5
- Registered claims: https://datatracker.ietf.org/doc/html/rfc7519#section-4.1
