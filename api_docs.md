Ethereum launcher API document

### send_transaction

Example
    
    // Request
    curl -X POST http://localhost:3000/launch_logs -d \
    '{
       "from": "0x31ebd457b999bf99759602f5ece5aa5033cb56b3",
       "to": "0x3eb06f432ae8f518a957852aa44776c234b4a84a",
       "value": "2000000000000000000",
       "data": [],
       "gas_price": "",
       "gas_limit": 30000,
       "item_type": "engine",
       "item_id": "7e36a266-1b32-4eaa-bc95-6d7b47451229",
       "isUrgent": false
     }'
    
    //Response
    '{
       "status": 0,
       "desc": "success",
       "data": {
         "data": {
           "hash": "0xeb3eeeb4e0b8c2e04920ccdcae2836c5a4172872a672045d444d33a56f825c3e",
           "item_type": "engine",
           "item_id": "7e36a266-1b32-4eaa-bc95-6d7b47451229",
           "status": 2,
           "gas_price": "13000000000",
           "gas_limit": "30000"
         }
       }
     }'

Request Data

| Parameter     | Description                                  | Required | Type   | Example    |
| ------------- | -------------------------------------------- | -------- | ------ | ---------- |
| **from**      | sender of the transaction                    | yes      | string | `0x31ebd457b999bf99759602f5ece5aa5033cb56b3` |
| **to**        | receiver of the transaction                  | yes      | string | `0x3eb06f432ae8f518a957852aa44776c234b4a84a` |
| **value**     | value of the transaction                     | yes      | string | `2000000000000000000` |
| **data**      | send data of the transaction                 | yes      | string | `[]` |
| **gas_price** | if empty, launcher will set a rational value | no       | string | `13000000000` |
| **gas_limit** | if empty, launcher will set a rational value | no       | int    | `30000` |
| **item_type** | id of the request system                     | yes      | string | `engine` |
| **item_id**   | id of the request log                        | yes      | string | `7e36a266-1b32-4eaa-bc95-6d7b47451229` |

Response Data
  
| Parameter     | Description               | Type   | Example    |
| ------------- | ------------------------- | ------ | ---------- |
| **hash**      | hash of the transaction   | string | `0xeb3eeeb4e0b8c2e04920ccdcae2836c5a4172872a672045d444d33a56f825c3e` |
| **status**    | status of the transaction | string | `2` |

Details

``status`` has 8 kind of value as follows:

| Code  | Status                | Description   |
| ----- | --------------------- | ------------- |
| **0** | CREATED               | status when launcher access a new tx|
| **1** | RETRIED               | a final status, when a tx was confirmed, other PENDING txs having the same nonce would turn into RETRIED |
| **2** | PENDING               | status when tx was successful sent |
| **3** | SUCCESS               | a final status, when the tx was confirmed on ethereum and successful|
| **4** | FAILED                | a final status, when the tx was confirmed on ethereum and failed |
| **5** | SIGN_FAILED           | a final status, when sign fail by pkm |
| **6** | SEND_FAILED           | a final status, when return error by sending tx to ethereum |
| **7** | ESTIMATED_GAS_FAILED  | a final status, when request gas limit is not enough |
	
### query_transaction

Example
    
    // Request
    curl -X GET http://localhost:3000/launch_logs -d \
    '{
       "hash": "0xeb3eeeb4e0b8c2e04920ccdcae2836c5a4172872a672045d444d33a56f825c3e",
       "item_id": "",
       "item_tyep": ""
     }'
    
    //Response
    '{
       "status": 0,
       "desc": "success",
       "data": {
         "data": [
           {
             "hash": "0xeb3eeeb4e0b8c2e04920ccdcae2836c5a4172872a672045d444d33a56f825c3e",
             "item_type": "engine",
             "item_id": "7e36a266-1b32-4eaa-bc95-6d7b47451229",
             "status": 3,
             "gas_price": "13000000000",
             "gas_used": 21000,
             "executed_at": 1578860943
           }
         ]
       }
     }'

Request Data

| Parameter       | Description                                 | Required | Type   | Example    |
| --------------- | ------------------------------------------- | -------- | ------ | ---------- |
| **hash**        | hash of the transaction you query           | no       | string | `0xeb3eeeb4e0b8c2e04920ccdcae2836c5a4172872a672045d444d33a56f825c3e` |
| **item_id**     | item_id of the transactions you query       | no       | string | `7e36a266-1b32-4eaa-bc95-6d7b47451229T` |
| **item_type**   | item_type of the transactions you query     | no       | string | `engine` |

Details

when you query logs by ``hash``, you will get only one log in response data. when you query logs by ``item_id`` and ``item_type``, you will get a list of logs.
this is because when the log you sent has been seed up, launcher will resend the log with a larger gas, and the response also contains all the resend logs.

Response Data

Return a list of logs data, the structure for every log item is same as ``send_transaction`` 
