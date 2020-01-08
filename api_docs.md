Ethereum launcher API document

### send_transaction

Example
    
    // Request
    curl -X GET http://localhost:3001/launch_logs -d \
    '{
        "from": "0x31ebd457b999bf99759602f5ece5aa5033cb56b3",
        "to": "0xd088fc0f4d5e68a3bb3d902b276ce45c598f858c",
        "value": "1000000000000000000",
        "data": [],
        "gas_price": "",
        "gas_limit": "",
        "item_type": "engine",
        "item_id": "7e36a266-1b32-4eaa-bc95-6d7b4745122e",
        "isUrgent": false,
     }'
    
    //Response
    '{
        "status": 0,
        "err_msg":"",
        "data": {
            "hash":"",
            "from": "0x31ebd457b999bf99759602f5ece5aa5033cb56b3",
            "to": "0xd088fc0f4d5e68a3bb3d902b276ce45c598f858c",
            "value": "1",
            "data": "",
            "gas_price": "",
            "gas_limit": "",
            "item_type": "",
            "item_id": "",
        },
     }'

Request Data

| Parameter     | Description                                  | Required | Type   | Example    |
| ------------- | -------------------------------------------- | -------- | ------ | ---------- |
| **from**      | sender of the transaction                    | yes      | string | `0x31ebd457b999bf99759602f5ece5aa5033cb56b3` |
| **to**        | receiver of the transaction                  | yes      | string | `0xd088fc0f4d5e68a3bb3d902b276ce45c598f858c` |
| **value**     | value of the transaction                     | yes      | string | `0` |
| **data**      | send data of the transaction                 | yes      | string | `0x` |
| **gas_price** | if empty, launcher will set a rational value | no       | string | `3` |
| **gas_limit** | if empty, launcher will set a rational value | no       | string | `25000` |
| **item_type** | id of the request system                     | yes      | string | `engine` |
| **item_id**   | id of the request log                        | yes      | string | `7e36a266-1b32-4eaa-bc95-6d7b4745122e` |

Response Data
  
| Parameter     | Description              | Type   | Example    |
| ------------- | ------------------------ | ------ | ---------- |
| **hash**      | hash of then transaction | string | `0x7572578cc4cf3e7c897811d48b65e01e2a5544a325acfdea831887d8a1a5703b` |

### query_transaction

Example
    
    // Request
    curl -X GET http://localhost:3001/launch_logs -d \
    '{
       "hash": "0x7572578cc4cf3e7c897811d48b65e01e7a5544a325acfdea831887d8a1a5703b",
       "item_id": "",
       "item_tyep": "",
     }'
    
    //Response
    '{
       "status": 0,
       "err_msg":"",
       "data":[
          {
             "hash":"",
             "from": "0x31ebd457b999bf99759602f5ece5aa5033cb56b3",
             "to": "0xd088fc0f4d5e68a3bb3d902b276ce45c598f858c",
             "value": "1",
             "data": "",
             "gas_price": "",
             "gas_limit": "",
             "item_type": "",
             "item_id": "",
          }
       ],
     }'

Request Data

| Parameter     | Description                             | Required | Type   | Example    |
| ------------- | --------------------------------------- | -------- | ------ | ---------- |
| **hash**      | hash of the transaction you query       | no       | string | `ETH-USDT` |
| **item_id**   | item_id of the transactions you query   | no       | string | `ETH-USDT` |
| **item_type** | item_type of the transactions you query | no       | string | `ETH-USDT` |

Details
when you query logs by ``hash``, you will get only one log in response data. when you query logs by ``item_id`` and ``item_type``, you will get a list of logs.
this is because when the log you sent has been seed up, launcher will resend the log with a larger gas, and the response also contains all the resend logs.

Response Data

Return a list of logs data, the structure for every log item is same as ``send_transaction`` 
