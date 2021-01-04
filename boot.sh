#!/bin/bash -ex

# export PION_LOG_TRACE=all

# Uncomment to test with VP8
# ffmpeg -re -f lavfi -i "testsrc=size=1920x1080:rate=30" -an -c:v libvpx -pix_fmt yuv420p -f ivf - | ./pion-ivf-server

# Uncomment to test with H264
ffmpeg -f lavfi -i "testsrc=size=1920x1080:rate=30" -an -c:v libx264 -profile:v baseline -bsf:v h264_mp4toannexb -pix_fmt yuv420p -f h264 - | ./pion-h264-server
