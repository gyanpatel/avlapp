FROM golang:latest
RUN mkdir /build
WORKDIR /build/
RUN cd /build && git clone https://github.com/gyanpatel/avlapp.git
RUN cd /build/avlapp && go mod tidy
RUN cd /build/avlapp && go build .
EXPOSE 8085
ENTRYPOINT [ "/build/avlapp/bccavl" ]