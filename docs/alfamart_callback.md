# Notification callback from alfamart service

- Endpoint: `/alfamart_payment_callback`
- HTTP Method: `POST`
- Request Header:
  - Accept: `application/json`
  - Content-type: `application/json`
- Request Body:
  ```JSON
  {
    "payment_id": "123123123",
    "payment_code": "XYZ123",
    "amount": 50000,
    "paid_at": "2020-10-17T07:41:33.866Z",
    "external_id": "order-123",
    "customer_id": 1,
  }
  ```
