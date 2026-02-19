FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/app ./cmd/app

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/app /app
COPY configs /configs
ENV APP_PROFILE=prod
EXPOSE 8080
ENTRYPOINT ["/app"]