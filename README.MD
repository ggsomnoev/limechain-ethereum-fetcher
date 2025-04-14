# EthereumFetcher

## Description
Dear reader, I hope this README finds you well. Lets begin!

EthereumFetcher is an app that fetches transaction data from a given ethereum node. The app provides:
 * DB caching of previously requested transaction data
 * Previous user transaction request data. Authorization is required.

## Architecture
Decided to structure the project using [clean architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html) in order to benefit of:
* easier testing and mocking
* easier future scaling
* easier future swap between components
* better structure, separation of concern, maintainability, reusability, etc.

Also I am using [Counterfeiter](https://github.com/maxbrunsfeld/counterfeiter) to generate fakes. [Ginkgo](https://onsi.github.io/ginkgo/) and [Gomega](https://onsi.github.io/gomega/) for testing (Why? I get better structure, readability, etc. Also I am most familiar with it). [Migrate](https://pkg.go.dev/github.com/golang-migrate/migrate/v4) for dealing with migrations. [Echo](https://echo.labstack.com/) web server. It is lightweight and easy to use.

Useful links:
* [C4 Diagram](./docs/c4diagram.md)

## How to run the server

To run the server using docker:

```bash
make run-docker
```

If you want to manually run the migrations:

```bash
make migrate-down
make migrate-up
```

## How to run the tests

To run the unit tests:

```bash
make test
```

## Example requests and responses

### To authenticate and get a token

There is a predefined set of accepted users, you can find them in the initial migration. Example request:

```bash
curl -X POST http://127.0.0.1:8080/lime/authenticate \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "password": "alice"
  }'
```

Example response:

```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDQ3MTk1OTMsInN1YiI6ImFsaWNlIn0.9kdexB3sdnIoNY2YKdDgMB-p7t2yk2JQvPE7Cizh9Rg"
}
```

You can use this token as an authorization header when accessing any other available endpoint to enable the user transaction tracking. Example request: 

```bash
curl -H "Authorization: Bearer previously.generated.token" -X GET http://localhost:8080/lime/eth?transactionHashes=0x1a2b3c&transactionHashes=0x3d4e5f
```

### Fetch & Store Ethereum Transactions

Example request:

```bash
curl -X GET http://127.0.0.1:8080/lime/eth?transactionHashes=0xfc2b3b6db38a51db3b9cb95de29b719de8deb99630626e4b4b99df056ffb7f2e&transactionHashes=0x48603f7adff7fbfc2a10b22a6710331ee68f2e4d1cd73a584d57c8821df79356&transactionHashes=0xcbc920e7bb89cbcb540a469a16226bf1057825283ab8eac3f45d00811eef8a64&transactionHashes=0x6d604ffc644a282fca8cb8e778e1e3f8245d8bd1d49326e3016a3c878ba0cbbd
```

### Fetch & Store Ethereum Transactions by RLP

Example request:

```bash
curl -X GET http://127.0.0.1:8080/lime/eth/f90110b842307866633262336236646233386135316462336239636239356465323962373139646538646562393936333036323665346234623939646630353666666237663265b842307834383630336637616466663766626663326131306232326136373130333331656536386632653464316364373361353834643537633838323164663739333536b842307863626339323065376262383963626362353430613436396131363232366266313035373832353238336162386561633366343564303038313165656638613634b842307836643630346666633634346132383266636138636238653737386531653366383234356438626431643439333236653330313661336338373862613063626264
```

### Get All Stored Transactions

Example request:

```bash
curl -X GET http://localhost:8080/lime/all
```

### Get All Stored User Transactions

Example request:

```bash
curl -H "Authorization: Bearer previously.generated.token" -X GET http://localhost:8080/lime/my
```

### Expected example response:

```json
{
  "transactions": [
    {
      "transactionHash": "0x48603f7adff7fbfc2a10b22a6710331ee68f2e4d1cd73a584d57c8821df79356",
      "transactionStatus": 1,
      "blockHash": "0x61914f9b5d11dcf30b943f9b6adf4d1c965f31de9157094ec2c51714cb505577",
      "blockNumber": 5703601,
      "from": "0x1fc35B79FB11Ea7D4532dA128DfA9Db573C51b09",
      "to": "0xAa449E0226B45D2044B1f721D04001fDe02ABb08",
      "contractAddress": "",
      "logsCount": 0,
      "input": "0x",
      "value": "500000000000000000"
    },
    {
      "transactionHash": "0xcbc920e7bb89cbcb540a469a16226bf1057825283ab8eac3f45d00811eef8a64",
      "transactionStatus": 1,
      "blockHash": "0xc5a3664f031da2458646a01e18e6957fd1f43715524d94b7336a004b5635837d",
      "blockNumber": 5702816,
      "from": "0xd5e6f34bBd4251195c03e7Bf3660677Ed2315f70",
      "to": "0x4c16D8C078eF6B56700C1BE19a336915962df072",
      "contractAddress": "",
      "logsCount": 1,
      "input": "0x6a627842000000000000000000000000d5e6f34bbd4251195c03e7bf3660677ed2315f70",
      "value": "0"
    }
  ]
}
```

## Future work:

Nice to add:
* linter config and fixes (it should have been from the start, yes)
* plan how to scale the app up - different enviroments, ci/cd, tf config, etc.
* healthcheck endpoint
* metrics and tracing
* parallel transaction processing
* modify the `eth/my` endpoint to support getting user transaction by date(optional)
* more..