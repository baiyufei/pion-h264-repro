This is a reduced test case of an issue I've been encountering with pion and H264 streams. 

## The problem

When reading VP8 video from a file descriptor and sending it to a client over WebRTC, the video starts instantly.

When reading H264 video from a file descriptor and sending it to a client over WebRTC, the video doesn't start for 8 seconds (and starts 8 seconds in).

## How to test it

`./build.sh` will build a docker image and `./run.sh` will run it. Or if you already have go and ffmpeg installed, then just compile the .go files and use `./boot.sh`.

Then go to http://localhost:8080 to view the page. You should see the ffmpeg testsrc pattern.
