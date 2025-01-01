# Online-Boutique


Online Boutique is a cloud-first microservices demo application. The application is a web-based e-commerce app where users can browse items, add them to the cart, and purchase them.

Adapted from https://github.com/GoogleCloudPlatform/microservices-demo/tree/main (all of the sevices that weren't written in Go ahve been ported to Go.)

## Architecture

**Online Boutique** is composed of 11 microservices written in different
languages that talk to each other over gRPC.

[![Architecture of
microservices](./architecture-diagram.png)](./architecture-diagram.png)

Find **Protocol Buffers Descriptions** at the [`./protos` directory](/protos).


## Build Docker Images and Push them to DockerHub

```bash
sudo bash build_images.sh # you need to change the username and run docker login
```

## Run Bookinfo Applicaton

```bash
kubectl apply -Rf ./kubernetes/apply
kubectl get pods
```