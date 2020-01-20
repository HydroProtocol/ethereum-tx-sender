# Ethereum tx sender

A platform focused on sending transactions to Ethereum. 
  
  For the developer or team of an Ethereum dapp, sending transactions to the Ethereum blockchain must be an essential link.
Which can not avoid to deal with some tedious problems. For example, what to do if the sent transaction was blocked, the nonce is wrong, the amount of gas used is relatively reasonable, and the status of the transaction is synchronized in time.
  
  The purpose of this system is to help developers solve these problems and allow developers to focus more on business logic.

## Features:

- Accept the encapsulated Ethereum transaction request and send it to Ethereum, and synchronize the transaction status.
- Query the status of sent transactions.
- Internal nonce management, set nonce automatically and orderly.
- Speed up transactions. When transactions are congested, gas will be resent according to network adjustments.
- Automatic calculation of gas limit.
- Get gas price in real time.

## Install & start ethereum-tx-sender

### Prerequisites

The only required software that you must have installed are `docker` and `docker-compose`.

If you don't already have them installed, you can follow [this link](https://docs.docker.com/compose/install/) to install them (free).

### Start a local ethereum-tx-sender

1.  **Clone this repo**

        git clone https://github.com/hydroprotocol/ethereum-tx-sender.git

1.  **Change your working directory**

        cd ethereum-tx-sender

1.  **Build and start ethereum-launch**

        docker-compose -f configs/docker-compose.yaml pull && docker-compose -f configs/docker-compose.yaml up -d

    This step may takes a few minutes.
    When complete, it will start all necessary services.

    It will use ports `3000`, `5432` and `8545` on your computer. Please make sure theses ports are available.

1.  **Check out your ethereum-tx-sender**

    Open http://localhost:3000/ on your browser to access ethereum-tx-sender


### Send transaction by ethereum-tx-sender

ethereum-tx-sender provide two interface, ``send_transaction`` and ``query_transaction`` for details see [api doc](docs/api.md).

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
        curl -X GET http://localhost:3000/launch_logs -d \
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

## Configurations
You can configure ```DATABASE_URL```, ```ETHEREUM_NODE_URL```, ```PRIVATE_KEYS``` environment variables as you need. for details see [envs.md](docs/envs.md).

## Notice
- For one address, it is better not to use the sender to send transactions and send transactions elsewhere, as this will cause problems with nonce
- It is best for developers to implement a set of pkm interfaces themselves, of course, a local simplified version of pmk is provided in the project

## What next
- A management background
- sdk for api

## Contributing

1. Fork it (<https://github.com/HydroProtocol/ethereum-tx-sender/fork>)
2. Create your feature branch (`git checkout -b feature/fooBar`)
3. Commit your changes (`git commit -am 'Add some fooBar'`)
4. Push to the branch (`git push origin feature/fooBar`)
5. Create a new Pull Request

## License
This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details
