#!/bin/bash

/tunnelclient
traceroute "$DEBUG_HOST"
ping -c 10 "$DEBUG_HOST"
curl -X POST "http://$DEBUG_HOST/placeholder" -v