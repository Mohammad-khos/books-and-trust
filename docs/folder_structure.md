.
├── Makefile
├── docker-compose.yml
├── docs
│   ├── README.md
│   └── folder_structure.md
├── go.mod
├── go.sum
├── infrastructure
│   └── deployment
│       ├── api-gateway.Dockerfile
│       ├── trust-service.Dockerfile
│       └── user-service.Dockerfile
├── proto
│   ├── trust.proto
│   └── user.proto
├── script
├── services
│   ├── api-gateway
│   │   ├── cmd
│   │   │   └── api
│   │   │       ├── main.go
│   │   │       └── mount.go
│   │   └── internal
│   │       └── handler
│   │           ├── grpc
│   │           ├── http
│   │           │   └── heath_check.go
│   │           └── middleware
│   │               └── csp_middleware.go
│   ├── trust-service
│   │   ├── cmd
│   │   │   └── api
│   │   │       └── main.go
│   │   └── internal
│   └── user-service
│       ├── Makefile
│       ├── cmd
│       │   └── api
│       │       └── main.go
│       ├── config
│       │   └── config.go
│       ├── internal
│       │   ├── domain
│       │   │   ├── repository.go
│       │   │   └── user.go
│       │   ├── handler
│       │   │   └── grpc
│       │   │       └── grpc_handler.go
│       │   ├── infra
│       │   │   └── repo
│       │   │       └── sql_repo.go
│       │   └── service
│       │       └── service.go
│       └── migrations
├── shared
│   ├── contracts
│   │   └── http.go
│   ├── db
│   │   └── postgres_db.go
│   ├── env
│   │   └── env.go
│   ├── json
│   │   └── json.go
│   ├── log
│   │   └── zap_logger.go
│   ├── migration
│   │   └── migration.go
│   ├── proto
│   │   ├── trust
│   │   │   ├── trust.pb.go
│   │   │   └── trust_grpc.pb.go
│   │   └── user
│   │       ├── user.pb.go
│   │       └── user_grpc.pb.go
│   └── validation
│       └── validate.go
└── web

43 directories, 35 files
