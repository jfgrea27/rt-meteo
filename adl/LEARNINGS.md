# Learnings

This file includes a stream of learnings from building this project.

## Learning 1: north vs. south & east vs. west traffic

North/south traffic controls external world to inside the cluster

East/west traffic controls inside the cluster service communication.

An ingress controller is responsible for routing external traffic to the right pod in the cluster. Example technologies include NGINX, Traefik, AWS Application Load Balancer.

A service mesh controls mTLS, retries, circuit breaking, observability, used for east/west communication. Example technologies include istio, linkerd.
