FROM golang:1.20 as build
ENV CGO_ENABLED=0
WORKDIR /workspace/src
COPY . .
RUN go get .
RUN go mod tidy
RUN go build -asmflags -trimpath -o /vault-audit .

FROM gcr.io/distroless/static AS final
COPY --from=build /vault-audit /vault-audit
ENTRYPOINT ["/vault-audit"]
