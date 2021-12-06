#!/bin/sh
for image in Dockerfiles/*; do
    tag="ghcr.io/slntopp/nocloud-tunnel-mesh/$(basename $image):latest"
    docker push $tag
done