## Notification service

This service can receive a payment notification from the Alfamart service and forward the payload to the involving customer.

### Endpoints

1. `POST` /alfamart_payment_callback
2. `POST` /login
3. `POST` /register
4. `POST` /callback_url

### Run the app with docker-compose

Services: app, postgres, redis

```sh
docker-compose up
```

#### Run all tests

```sh
go test ./...
```
