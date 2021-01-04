#!/bin/bash -e

close_it_down(){
    docker kill "$(docker ps -q)"
}

trap "close_it_down" SIGINT

docker run --rm --network host -t pion-h264-repro &

wait