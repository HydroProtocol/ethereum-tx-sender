# Ethereum tx sender

An interface for creating and sending transactions on the [Ethereum Blockchain](https://ethereum.org/). 
  
  Ethereum dApps are applications which use the Ethereum Blockchain. These dApps need to automatically create and send blockchain transactions as an essential part of the user experience.
Dynamically creating and sending blockchain transactions consistently is actually quite complex, with a number of challenging nuances. For example, sent transactions can be blocked, the nonce supplied by the user's wallet could be wrong, the gas settings need to be relatively reasonable, etc. These are essential things that every dApp developer needs to address.
  
  The purpose of this system is to provide a ready-to-use interface for creating and sending transactions on Ethereum, allowing developers to focus on their core business logic rather than deal with blockchain complications.

## Features:

- Accepts encapsulated Ethereum transaction requests and send it to the blockchain.
- Synchronizes transaction status.
- Query the status of sent transactions.
- Internal nonce management (automatic nonce settings)
- Speed up transactions. When transactions are congested, tx's can be automatically resubmitted with higher gas prices.
- Automatic calculation of the transaction's appropriate gas limit.
- Get an appropriate gas price in real time.

## Install & start ethereum-tx-sender

### Prerequisites

The only required software to run the ethereum-tx-sender are `docker` and `docker-compose`.

If you don't already have them installed, you can follow [this link](https://docs.docker.com/compose/install/) to install them (free).

### Start a local ethereum-tx-sender

1.  **Clone this repo**

        git clone https://github.com/hydroprotocol/ethereum-tx-sender.git

1.  **Change your working directory**

        cd ethereum-tx-sender

1.  **Build and start**

        docker-compose -f configs/docker-compose.yaml pull && docker-compose -f configs/docker-compose.yaml up -d

    This step may takes a few minutes.
    When complete, it will start all the necessary services.

    It will use ports `3000`, `5432` and `8545` on your computer. Please make sure these ports are available.

1.  **Check out your ethereum-tx-sender**

    Open http://localhost:3000/ on your browser to access your ethereum-tx-sender


### Send transaction by ethereum-tx-sender

ethereum-tx-sender comes with two interfaces: ``send_transaction`` and ``query_transaction``. For more details see [api doc](docs/api.md).

1.   **sending a transaction**

         // send request 
         curl -X POST http://localhost:3000/launch_logs -d \
         '{
           "from": "0x31ebd457b999bf99759602f5ece5aa5033cb56b3",
           "to": "0x3eb06f432ae8f518a957852aa44776c234b4a84a",
           "value": "2000000000000000000",
           "data": [],
           "item_type": "engine",
           "item_id": "7e36a266-1b32-4eaa-bc95-6d7b47451221"
         }'

         // response
         {
           "status": 0,
           "desc": "success",
           "data": {
             "data": {
               "hash": "0x20f1b07522f385e84e8f75f99b3e2d7b22d915bc735824336d662bcbae7e542a",
               "item_type": "engine",
               "item_id": "7e36a266-1b32-4eaa-bc95-6d7b47451221",
               "status": 2,
               "gas_price": "13000000000",
               "gas_limit": "25200"
             }
           }
         }
1.  **querying a transaction result**

        // send request
        curl -X GET http://localhost:3000/launch_logs -d \
        '{
            "hash": "0x20f1b07522f385e84e8f75f99b3e2d7b22d915bc735824336d662bcbae7e542a",
         }'

        // response
        {
          "status": 0,
          "desc": "success",
          "data": {
            "data": [
              {
                "hash": "0x20f1b07522f385e84e8f75f99b3e2d7b22d915bc735824336d662bcbae7e542a",
                "item_type": "engine",
                "item_id": "7e36a266-1b32-4eaa-bc95-6d7b47451221",
                "status": 3,
                "gas_price": "13000000000",
                "gas_used": 21000,
                "executed_at": 1579542180
              }
            ]
          }
        }

## Configurations
You can configure ```DATABASE_URL```, ```ETHEREUM_NODE_URL```, ```PRIVATE_KEYS``` environment variables as you need. For details see [envs.md](docs/envs.md).

## Notice
- Nonce problems can occur if you send transactions with an address using other tx services in addition to this project. Once you start using this project for a given address, it's best to not use other transaction sending services for that same address, or you may encounter a nonce issue.
- It is best for developers to implement a set of private key managemenet (pkm) interfaces themselves. However, a local simplified version of pkm is provided in this project for ease for use.


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
