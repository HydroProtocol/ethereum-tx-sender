## Environment variables
Before starting ethereum-sender, you must first configure the following environment variables.
You can also modify them in file `docker-compose.yaml` if you start it by ``docker-compose``.

#### DATABASE_URL
database to store launch logs. e.g. ```postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable```  
#### ETHEREUM_NODE_URL
a ethereum node to send transaction and query transaction receipt. e.g. ```http://localhost:8545```  
#### PRIVATE_KEYS
to sign the transactions. e.g. ```0xb7a0c9d2786fc4dd080ea5d619d36771aeb0c8c26c290afd3451b92ba2b7bc2c``` 
