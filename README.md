# Ethereum Launcher

A platform focused on sending transactions to Ethereum. 
  
  For the developer or team of an Ethereum dapp, sending transactions to the Ethereum blockchain must be an essential link.
Which can not avoid to deal with some tedious problems. For example, what to do if the sent transaction was blocked, the nonce is wrong, the amount of gas used is relatively reasonable, and the status of the transaction is synchronized in time.
  
  The purpose of this system is to help developers solve these problems and allow developers to focus more on business logic.

##Features:

- Accept the encapsulated Ethereum transaction request and send it to Ethereum, and synchronize the transaction status.
- Query the status of sent transactions.
- Internal nonce management, set nonce automatically and orderly.
- Speed up transactions. When transactions are congested, gas will be resent according to network adjustments.
- Automatic calculation of gas limit.
- Get gas price in real time.

## Install & start etherum-launcher

### Prerequisites

The only required software that you must have installed are `docker` and `docker-compose`.

If you don't already have them installed, you can follow [this link](https://docs.docker.com/compose/install/) to install them (free).

### Start a local etherum-launcher

1.  **Clone this repo**

        git clone https://github.com/hydroprotocol/ethereum-launcher.git

1.  **Change your working directory**

        cd ethereum-launcher

1.  **Build and start ethereum-launch**

        docker-compose pull && docker-compose up -d

    This step may takes a few minutes.
    When complete, it will start all necessary services.

    It will use ports `3000`, `6379` and `8545` on your computer. Please make sure theses ports are available.

1.  **Check out your ethereum-launcher**

    Open http://localhost:3000/ on your browser to access ethereum-launcher


### Send transaction by ethereum-launcher

ethereum-launcher provide two interface, ``send_transaction`` and ``query_transaction`` for details see [api doc](./api_docs.md).

1.   **send a transaction**

         // send request 
         curl -X POST http://localhost:3000/launch_logs -d \
         '{
            "from": "0x31ebd457b999bf99759602f5ece5aa5033cb56b3",
            "to": "0xd088fc0f4d5e68a3bb3d902b276ce45c598f858c",
            "value": "1000000000000000000",
            "data": [],
            "item_type": "engine",
            "item_id": "7e36a266-1b32-4eaa-bc95-6d7b4745122e",
         }'

         // response
         {
            "status": 0,
            "err_msg":"",
            "data":"",
         }
1.  **query transaction result**

        // send request
        curl -X GET http://localhost:3001/launch_logs -d \
        '{
            "hash": "0x7572578cc4cf3e7c897811d48b65e01e7a5544a325acfdea831887d8a1a5703b",
            "item_id": "7e36a266-1b32-4eaa-bc95-6d7b4745122e",
            "item_tyep": "engine",
         }'

        // response
        {
           "status":"",
           "err_msg":"",
           "data":"",
        }

##Configurations
You can modify the following environment variables in file `docker-compose.yaml`

   ``DATABASE_URL`` - database to store launch logs   
   ``ETHEREUM_NODE_URL`` - a ethereum node to send transaction and query transaction receipt  
   ``MAX_GAS_PRICE_FOR_RETRY`` - the max value of gas price, to prevent infinite gas price increases     
   ``RETRY_PENDING_SECONDS_THRESHOLD`` - the interval time to speed up the pending transaction  
   ``RETRY_PENDING_SECONDS_THRESHOLD_FOR_URGENT`` - the interval time to speed up the urgent pending transaction  
   ``PRIVATE_KEYS`` - to sign the transactions  

##Notice
- For one address, it is better not to use the launcher to send transactions and send transactions elsewhere, as this will cause problems with nonce
- It is best for developers to implement a set of pkm interfaces themselves, of course, a local simplified version of pmk is provided in the project

##What next
- A management background
- sdk for api

##License
This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details
