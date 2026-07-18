# Google Takeout Mbox Parser (DDD Hexagonal Architecture)

This repository contains a Go implementation of an mbox file parser, designed following **Domain-Driven Design (DDD)** and **Hexagonal (Ports and Adapters) Architecture** principles. It reads large Google Takeout `.mbox` mail exports sequentially and outputs structured email data in JSON Lines (`.jsonl`) format.

## Architecture

The project is structured into three distinct layers, ensuring that the core business domain logic is decoupled from input/output drivers and databases:

1. **Domain Layer**: Contains the core model representation (`Email`, `Attachment`) and domain-level errors. It has no external dependencies.
2. **Ports Layer**: Declares interfaces (the boundaries) defining how the outside world interacts with the domain (`inbound` / driving ports) and how the domain interacts with the outside world (`outbound` / driven ports).
3. **Application & Adapters Layer**:
   - **Application**: Implements the use cases (inbound ports) and orchestrates the domain models and outbound ports.
   - **Adapters**:
     - *Inbound (Driving)*: The Command Line Interface (CLI) program (`cmd/cli/main.go`) which parses inputs, executes the use cases, and prints live execution progress.
     - *Outbound (Driven)*: The `.mbox` parser (`adapters/outbound/mbox`) which streams files and parses complex MIME/multipart mail bodies, decodes headers (RFC 2047), and extracts attachments. The repository adapter (`adapters/outbound/repository`) writes emails as JSON Lines to files or handles them in memory.

### Dependency Graph

```mermaid
graph TD
    %% Inbound Adapters
    subgraph Inbound Adapters (Driving)
        CLI[cmd/cli/main.go]
    end

    %% Inbound Ports
    subgraph Inbound Ports (Driving Ports)
        ImportUC[ports/inbound/ImportEmailsUseCase]
    end

    %% Domain Service & Core Entities
    subgraph Core Domain & Application
        Importer[application/services/EmailImporter]
        EmailModel[domain/model/Email]
        Errors[domain/errors]
    end

    %% Outbound Ports
    subgraph Outbound Ports (Driven Ports)
        MboxParserPort[ports/outbound/MboxParser]
        EmailRepoPort[ports/outbound/EmailRepository]
    end

    %% Outbound Adapters
    subgraph Outbound Adapters (Driven)
        MboxParserImpl[adapters/outbound/mbox/Parser]
        InMemoryRepo[adapters/outbound/repository/InMemoryRepository]
        JSONLinesRepo[adapters/outbound/repository/JSONLinesRepository]
    end

    %% Relationships
    CLI -->|drives| ImportUC
    Importer -->|implements| ImportUC
    Importer -->|uses| EmailModel
    Importer -->|uses| MboxParserPort
    Importer -->|uses| EmailRepoPort
    MboxParserImpl -.->|implements| MboxParserPort
    InMemoryRepo -.->|implements| EmailRepoPort
    JSONLinesRepo -.->|implements| EmailRepoPort
```

---

## File Structure

- [cmd/cli/main.go](file:///Volumes/External/Code/takeout_services/cmd/cli/main.go): The entry point for the CLI. Sets up dependencies and coordinates the use case execution.
- [internal/domain/model/email.go](file:///Volumes/External/Code/takeout_services/internal/domain/model/email.go): The core Domain Entity `Email`.
- [internal/domain/errors.go](file:///Volumes/External/Code/takeout_services/internal/domain/errors.go): Domain-specific error definitions.
- [internal/ports/inbound/import_emails.go](file:///Volumes/External/Code/takeout_services/internal/ports/inbound/import_emails.go): Driving port interface (`ImportEmailsUseCase`).
- [internal/ports/outbound/email_repo.go](file:///Volumes/External/Code/takeout_services/internal/ports/outbound/email_repo.go): Driven port interface (`EmailRepository`).
- [internal/ports/outbound/mbox_parser.go](file:///Volumes/External/Code/takeout_services/internal/ports/outbound/mbox_parser.go): Driven port interface (`MboxParser`).
- [internal/application/services/email_importer.go](file:///Volumes/External/Code/takeout_services/internal/application/services/email_importer.go): Application Service coordinating parser and repo ports.
- [internal/adapters/outbound/mbox/parser.go](file:///Volumes/External/Code/takeout_services/internal/adapters/outbound/mbox/parser.go): Mbox parser implementation. Supports MIME multipart, Base64/Quoted-Printable decoding, header decodings, and attachment parsing.
- [internal/adapters/outbound/repository/repository.go](file:///Volumes/External/Code/takeout_services/internal/adapters/outbound/repository/repository.go): In-memory and JSON Lines output file adapters.

---

## Getting Started

### 1. Run Unit & Integration Tests
Verify the code correctness by running the test suite:
```bash
go test -v ./...
```

### 2. Build the CLI Tool
Build the executable binary:
```bash
go build -o takeout-parser ./cmd/cli
```

### 3. Parse Google Takeout Mail
Run the compiled binary by passing your input `.mbox` file:
```bash
./takeout-parser --input "Takeout/Почта/Вся почта, включая _Спам_ и _Корзину_.mbox" --output emails.jsonl
```

- `--input`: Path to the input `.mbox` file.
- `--output`: Path to write the output JSON Lines file (defaults to `emails.jsonl`).
- `--verbose`: Enable/disable stdout progress logs (defaults to `true`).
