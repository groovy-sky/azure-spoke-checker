# Use the official Golang image to create a build artifact  
FROM golang:alpine3.19 as builder  

ENV WEB web-app

# Copy go mod and sum files  
COPY  ${WEB}/* /opt

# Set working directory  
WORKDIR /opt
  
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed  
RUN go mod download  
  
# Build the Go app  
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .  && chmod +x main
  
# Use the official Hashicorp Terraform image as a base  
FROM hashicorp/terraform:1.8.1 as tf

FROM alpine:3.19

ENV TF tf-files
ENV WEB web-app
  
COPY ${TF} ${TF}

WORKDIR ${TF}/workflow

COPY --from=tf /bin/terraform /bin/terraform
COPY --from=builder /opt/main /bin/${WEB}
 
# Initialize Terraform  
RUN terraform init  && ls /bin/
  
# Run the executable  
CMD ${WEB}