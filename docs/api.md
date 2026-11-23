# Orders API

The edge service exposes the Orders API under `/api/v1` at
`http://localhost:8080`. It strips that prefix when forwarding to Orders. The
Orders service is also available directly at `http://localhost:8081` with the
same routes minus `/api/v1`.

These endpoints are for the reference implementation and do not require
authentication. Do not expose them to untrusted networks.

## Create an order

`POST /api/v1/orders` (edge) or `POST /orders` (Orders service)

Request body:

```json
{
  "customerId": "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
  "items": [
    {
      "productId": "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
      "quantity": 2
    }
  ]
}
```

`customerId` and each `productId` must be non-zero UUIDs. At least one item is
required, and each item's `quantity` must be between 1 and 1,000 inclusive.
The request must contain one valid JSON value.

On success, the service returns `201 Created` with the created order:

```json
{
  "orderId": "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
  "customerId": "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
  "items": [
    {
      "productId": "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
      "quantity": 2
    }
  ],
  "status": "pending",
  "createdAt": "2026-01-02T03:04:05Z"
}
```

Invalid JSON or order data returns `400 Bad Request`. An order-ID conflict
returns `409 Conflict`. Unexpected failures return `500 Internal Server Error`.

## Retrieve an order

`GET /api/v1/orders/{orderId}` (edge) or `GET /orders/{orderId}` (Orders
service)

`orderId` must be a UUID. A found order returns `200 OK` with the same response
shape as order creation. An invalid UUID returns `400 Bad Request`; an unknown
order returns `404 Not Found`.

## Health and readiness

Both services expose `GET /health` and `GET /ready`. Health reports that the
process is responding. Orders readiness checks its PostgreSQL connection and
returns `503 Service Unavailable` when the database is unavailable. Edge
readiness reports that the edge process is responding; it does not verify
upstream Orders availability.

## Error format

API errors use `application/problem+json` with an RFC 9457-style body:

```json
{
  "title": "Bad Request",
  "status": 400,
  "detail": "invalid request body"
}
```
