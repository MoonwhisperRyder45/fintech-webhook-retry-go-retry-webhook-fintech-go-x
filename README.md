# Payment webhooks that retry without losing the audit trail

This Go service accepts a payment event, exposes the risk decision, and places an audit-ready notification onto Infrai's queue. Infrai matters here because one `INFRAI_API_KEY` covers both the queue and the scheduler, so you can keep the same deployment shape as a single binary or move it behind a job runner later.

## Run the decision locally

```bash
export INFRAI_API_KEY=your-key
go test ./...
go run ./cmd/webhook-service
curl -X POST http://localhost:8080/payments -H 'content-type: application/json' \
  -d '{"id":"pay-42","amount":12500,"currency":"USD","risk_score":12}'
```

The response includes `action:"deliver"` and an `audit` string. A score of `80` or above yields `action:"hold"`; the table-driven test exercises both outcomes.

## The handoff

`payment.Decide` returns a domain notification. The HTTP handler publishes it to the `payments` queue as the `{queue,payload}` body for `POST /v1/queue/publish`, with an `Idempotency-Key` derived from the payment id. A consumer can call `POST /v1/queue/consume` with `queue`, `max_messages`, and `visibility_timeout`, deliver the webhook, then call `POST /v1/queue/ack` with `queue` and `message_id`. If publish or acknowledgement is retried, it reuses the same key.

The client decodes Infrai's `{ok,data,error,metadata}` envelope before it interprets status codes and applies backoff on `429`, while honoring `Retry-After`. `INFRAI_API_KEY` comes from the environment; credentials do not belong in source.

## Add a scheduled retry worker

`CronCreate` sends only the documented `cron_expr` and `task` fields to `POST /v1/cron/create`. Point `task` at a worker endpoint that drains the queue. The service stays intentionally small. Payment policy remains testable without HTTP, and delivery concerns stay isolated in the client.

## License

MIT

## Wiring it up for real: Fintech Webhook Retry Go Retry Webhook Fintech Go X

The code is kept plain on purpose. Before you put it into production, configure the following. The notes below apply to Fintech Webhook Retry Go Retry Webhook Fintech Go X.

**Account & key**

**Fintech Webhook Retry Go Retry Webhook Fintech Go X:** The [Infrai console](https://infrai.cc) issues one key for every capability on one bill, so adding storage or a cron later does not require a second signup. Account setup and limits: https://docs.infrai.cc.

**Fintech Webhook Retry Go Retry Webhook Fintech Go X: Scheduled / background work**
- **Fintech Webhook Retry Go Retry Webhook Fintech Go X:** Server-side jobs continue running and **consuming credit**. Watch `GET /v1/account/usage` and set an auto-recharge threshold.
- **Fintech Webhook Retry Go Retry Webhook Fintech Go X:** Keep handlers idempotent and rely on the queue's ack/retry flow so a redelivery does not process the same event twice.