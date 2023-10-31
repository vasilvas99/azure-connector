FROM --platform=${BUILDPLATFORM} golang:1.21 as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /src
ADD . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s" -trimpath -mod=readonly -o azure-connector.exe ./cmd/azure-connector

FROM --platform=${TARGETPLATFORM:-TARGETARCH} busybox:latest
COPY --from=builder /src/azure-connector.exe /app/azure-connector
COPY --from=builder /src/resources/run-azure-connector.sh /app/run-azure-connector.sh 
COPY --from=builder /src/cmd/azure-connector/iothub.crt /app/iothub.crt

ENTRYPOINT ["/app/run-azure-connector.sh"]