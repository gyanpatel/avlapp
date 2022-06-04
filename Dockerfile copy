FROM golang:latest
RUN mkdir /build
RUN cp -rf * /build/
WORKDIR /build/bccavl/
RUN go mod tidy
RUN go build .
EXPOSE 8085
ENTRYPOINT [ "/build/bccavl" ]