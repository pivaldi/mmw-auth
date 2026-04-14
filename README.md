# DDD & Hexagonal Architecture with Standard Go Project Layout

This Go module takes part of the [mmw](https://github.com/pivaldi/mmw) project that demonstrates the implementation of the [Go Modular Monolith White Paper](https://github.com/pivaldi/go-modular-monolith-white-paper).

This project is not usable independently of [mmw-todo](https://github.com/pivaldi/mmw-auth); the best way to test this project is to use directly the [Monolith Modular Worskpace](https://github.com/pivaldi/mmw).

## Overview

This repository includes a **working example implementation** that demonstrates how to apply Domain-Driven Design (DDD) and Hexagonal Architecture (Ports & Adapters) patterns using the [Standard Go Project Layout](https://github.com/golang-standards/project-layout/releases/latest)).

As is, this project takes part of [the `mmw` project](https://github.com/pivaldi/mmw) and should be use for now from this Monolith Modular Workspace.

### What's Included

The Todo API example provides:

- **Authentication process** register login user
- **Expose a private ValidateToken API** used by the [Todo project](https://github.com/pivaldi/mmw-todo)
- **Dual Protocol Support** - HTTP and gRPC from single protobuf definitions using [Buf Connect](https://connect.build)
- **Domain-Driven Design** - Rich domain model with aggregates, value objects, and domain events
- **Hexagonal Architecture** - Clear separation between domain, application, and infrastructure layers
- **PostgreSQL Persistence** - Repository pattern with database migrations
- **Comprehensive Testing** - Unit, integration, and API tests demonstrating testing strategies for each layer

### Quick Start

As is, this project takes part of [the `mmw` project](https://github.com/pivaldi/mmw) and should be use for now from this Monolith Modular Workspace.
