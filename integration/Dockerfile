FROM golang:1.22.5 AS builder

WORKDIR /app

COPY . ./

# Download Go modules
RUN go mod download

# Build the binary
RUN CGO_ENABLED=0 go build -o /gomi

FROM scratch

# Copy our static executable
COPY --from=builder /gomi ./
