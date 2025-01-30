# fetchpoints

The purpose of the API is to calculate the points accrued per purchase based on certain rules.

The API has two endpoints.

1. POST /receipts/process :  ```receipts/process```

   The JSON Payload of the receipt id necessary
   Once the request is received, the receipt is processed and points are generated and an ID is generated and sent 
   as a response to the client.


2. GET /receipts/{id}/points

   Upon receiving a request with a receipt id, the points that have been calculated when the
   receipt is processed are pulled per the provided receipt id and returned to the client.


Sample

POST Request Payload
```json
{
"retailer": "M&M Corner Market",
"purchaseDate": "2022-03-20",
"purchaseTime": "14:33",
"items": [
{
"shortDescription": "Gatorade",
"price": "2.25"
},{
"shortDescription": "Gatorade",
"price": "2.25"
},{
"shortDescription": "Gatorade",
"price": "2.25"
},{
"shortDescription": "Gatorade",
"price": "2.25"
}
],
"total": "9.00"
}
```

POST Response
```json
{
    "ID": 1
}
```


GET Request (for the above receipt) : ```/receipts/1/points```

Response (for the above receipt)
```json
{
    "points": 109
}
```