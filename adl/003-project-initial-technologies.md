# 003 Project initial technologies

As mentioned in [Project initial assumptions - Assumption 3](./001-project-initial-assumptions.md#assumption-3), this project doesn't necessarily aim to build the most cost-effective solution, but rather a real-world solution. This means we will favour using technologies that are in production today over simpler cheaper options. Some of these are listed here:

## Decision 1: Kubernetes over VM or vendor-specific offering

Since Kubernetes is so ubiquitous, and we might end up adding AI services, this project will favour Kubernetes over serverless or managed conrainer products (e.g. [AWS ECS](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/Welcome.html)). This will increase the burden on the maintainer of the system and increase the costs, but this is again just an exercise.Managing a VM instead also be unrealistic for production environments.

## Decision 2: Golang as the backend language of choice

Coding languages are just tools. This project could very much be written in Python or Rust or Typescript, Golang was just chosen for its portability and small container size footprint.

## Decision 3: Typescript and React for the UI.

Again, this is ubiquitous in the industry, so it has been chosen for this project.

## Decision 4: AWS as the cloud provider

AWS has the largest market share (28%), and so is the chosen provider ([ref](https://www.statista.com/chart/18819/worldwide-market-share-of-leading-cloud-infrastructure-service-providers/?srsltid=AfmBOorywAbylr1bvAYjfNa7oCEzJH41HtjuAR1z4h0N46W3qdtq5ip8)).

## Decision 5: Network routing technologies

This project will use the following network routing technologies:

- [AWS Application Load Balancer](https://aws.amazon.com/elasticloadbalancing/application-load-balancer/) for public TLS termination, DDoS protection, firewall.
- [Traefik](https://traefik.io/) as the ingress controller into the Kubernetes cluster. This will include routing to services from external, middleware and rate limiting.
- [linkerd](https://linkerd.io/) as the service mesh to provide secure mTLS, circuit breaking and observability.

ALB has poor service-to-service observability (so linkerd), limited routing flexibility (so traefik). Traefik has no edge protection, so ALB. Linkerd does not manage ingress into cluster.

## Decision 6: Communication protocols

Since the architecture for now is only has a single [Weather API](./002-project-initial-architecture.md#weather-api), a RESTfull API will be used between [Weather UI](./002-project-initial-architecture.md#weather-ui) and [Weather API](./002-project-initial-architecture.md#weather-api). For service-to-service communication, which might be implemented later, the project will favour using [gRPC](https://grpc.io/) over REST for backward compatibility and smaller payloads.

## Overall

The following diagram shows the current architecture with technologies for the system:

![Architectural diagram](../diagrams/diagrams-Network%20Architecture.drawio.png)
