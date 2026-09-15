FROM golang:1.27-alpine AS build

WORKDIR /src

COPY . .

RUN CGO_ENABLED=0 go build -o /xdomea-ui .

FROM scratch

WORKDIR /app

COPY --from=build /xdomea-ui /xdomea-ui
COPY --from=build /src/template ./template

EXPOSE 8080

ENTRYPOINT ["/xdomea-ui"]