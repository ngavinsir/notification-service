# Register customer

- Endpoint: `/register`
- HTTP Method: `POST`
- Request Header:
  - Accept: `application/json`
  - Content-type: `application/json`
- Request Body:
  ```JSON
  {
      "email": "string",
      "password": "string
  }
  ```
- Response Body:
  ```JSON
  {
      "id": "number",
      "email": "string"
  }
  ```

# Login customer

- Endpoint: `/login`
- HTTP Method: `POST`
- Request Header:
  - Accept: `application/json`
  - Content-type: `application/json`
- Request Body:
  ```JSON
  {
      "email": "string",
      "password": "string
  }
  ```
- Response Header:
  - Set-Cookie: `_gosession=ZXhhbXBsZTJAZXhhbXBsZS5jb20::bLRvn7yCSyS-bFx0dbYDOaxgq8_QJFrq`
- Response Body:
  ```JSON
  {
      "id": "number",
      "email": "string"
  }
  ```
