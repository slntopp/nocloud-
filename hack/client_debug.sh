#!/bin/bash

/tunnelclient
traceroute "$DEBUG_HOST"
ping -c 10 "$DEBUG_HOST"