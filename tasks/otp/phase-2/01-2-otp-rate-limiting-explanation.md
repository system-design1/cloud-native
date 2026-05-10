# OTP Send Flow, Active OTP Protection, and Rate Limiting

## Current Rate Limiting Logic

The current rate limiter only applies to:

```text
SendOTP
```

It does **not** apply to:

```text
VerifyOTP
```

At the moment, the implemented rate limiting scope is:

```text
tenant_id + phone
```

The Redis key format is:

```text
otp:rate:send:{tenant_id}:{phone}
```

Example:

```text
otp:rate:send:202:+989121234570
```

---

# Example Configuration

If the configuration is:

```env
OTP_SEND_RATE_LIMIT_ENABLED=true
OTP_SEND_RATE_LIMIT_MAX=2
OTP_SEND_RATE_LIMIT_WINDOW=10m
```

It means:

```text
For tenant=202 and phone=+989121234570,
within a 10-minute window,
a maximum of 2 successful OTP send operations are allowed.
The third accepted send attempt will be blocked with HTTP 429.
```

---

# Important Behavior: Active OTP Protection Runs Before Rate Limiting

The current `SendOTP` flow is approximately:

```text
1. validate request
2. tenant validation
3. check active OTP
4. if active OTP exists => return 429 OTP already active
5. if no active OTP exists => execute rate limiter check
6. if rate limit exceeded => return 429 OTP send rate limit exceeded
7. generate OTP
8. save OTP
9. send SMS
```

This is extremely important for understanding runtime behavior.

---

# Example Scenario: Multiple Immediate Requests

If a user sends three immediate requests using the same phone number, this is what usually happens.

## First Request

```text
No active OTP exists.
Rate limit counter becomes 1.
OTP send proceeds successfully.
```

Response:

```http
200 OK
```

---

## Second Immediate Request

Because the previous OTP is still active:

```text
Active OTP protection blocks the request.
The rate limiter is not executed at all.
```

Response:

```http
429 OTP already active
```

This does **not** consume rate limit quota.

---

# When Does the Rate Limiter Actually Trigger?

The rate limiter becomes visible only when the active OTP state is removed between send attempts.

This can happen when:

- the user successfully verifies the OTP
- the OTP expires naturally
- or the OTP Redis state is manually deleted during testing

Example manual cleanup:

```bash
redis-cli DEL 'otp:202:+989121234570'
```

---

# Example Runtime Sequence With MAX=2

Assume:

```env
OTP_SEND_RATE_LIMIT_MAX=2
```

Then:

## Send #1

```text
allowed
counter = 1
```

Delete OTP state manually:

```bash
redis-cli DEL 'otp:202:+989121234570'
```

---

## Send #2

```text
allowed
counter = 2
```

Delete OTP state manually again:

```bash
redis-cli DEL 'otp:202:+989121234570'
```

---

## Send #3

```text
blocked
counter = 3
3 > 2
```

Response:

```http
429 Too Many Requests
```

Message:

```text
OTP send rate limit exceeded
```

---

# What Rate Limiting Is Implemented Right Now?

Currently implemented:

```text
Per tenant + phone OTP send rate limiting
```

Meaning:

```text
A specific tenant can only send OTP a limited number of times
to a specific phone number during a configured time window.
```

---

# What Is NOT Implemented Yet?

The following rate limiting and abuse protection layers do not exist yet.

## Per Tenant Aggregate Limit

Not implemented.

Example:

```text
Tenant 202 cannot send more than 1000 OTPs per hour globally.
```

---

## Per IP Rate Limit

Not implemented.

Example:

```text
A single IP address cannot send more than 20 OTP requests per minute.
```

---

## Global Per Phone Limit

Not implemented.

Example:

```text
A single phone number cannot be spammed across multiple tenants.
```

---

## Provider Billing Protection

Not implemented.

Example:

```text
Protecting against excessive SMS provider billing or tenant abuse.
```

---

## Sliding Window / Token Bucket

Not implemented.

Current implementation uses:

```text
Fixed Window Rate Limiting
```

---

## Retry-After Header

Not implemented.

Currently the API does not return:

```text
Retry-After
```

or remaining retry time information.

---

## Phone Normalization / Hashing

Not implemented.

Currently the raw phone number is used directly inside Redis keys.

---

## Rate Limit Logging / Metrics

Not implemented.

Currently the system only returns HTTP responses and does not yet expose:

- rate limit metrics
- limiter observability
- limiter logging
- tracing for limiter events

---

# Final Summary

At the moment, there are two protection layers for `SendOTP`.

## 1. Active OTP Protection

```text
As long as an active OTP exists for tenant + phone,
resending is blocked.
```

---

## 2. Fixed-Window Rate Limiting

```text
If too many accepted OTP sends happen for the same tenant + phone
inside a configured time window,
additional sends are rejected.
```

---

# Important Final Note

This is not yet a complete production-grade rate limiting system.

However, for the current OTP phase, it is a correct and useful first abuse-protection layer.
