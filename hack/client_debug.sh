#!/bin/bash

/tunnelclient
curl -X POST "http://$DEBUG_HOST/i/am/debug/request" -v
traceroute "$DEBUG_HOST"
ping -c 10 "$DEBUG_HOST"