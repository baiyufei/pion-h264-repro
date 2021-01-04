FROM golang:1.15


# Supress warnings about missing front-end. As recommended at:
# http://stackoverflow.com/questions/22466255/is-it-possibe-to-answer-dialog-questions-when-installing-under-docker
ARG DEBIAN_FRONTEND=noninteractive
ENV GO111MODULE on

RUN apt-get update -y && \
    apt-get install -y ffmpeg && \
    apt-get clean autoclean && apt-get autoremove -y && rm -rf /var/lib/{apt,dpkg,cache,log}/

WORKDIR /home/

# COPY nv-tensorrt-repo-ubuntu1804-cuda10.0-trt5.1.5.0-ga-20190427_1-1_amd64.deb nv-tensorrt-repo-ubuntu1804-cuda10.0-trt5.1.5.0-ga-20190427_1-1_amd64.deb
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY ./*.go ./
RUN go build pion-ivf-server.go && go build pion-h264-server.go

COPY www www 
COPY ./*.sh ./

CMD ["./boot.sh"]