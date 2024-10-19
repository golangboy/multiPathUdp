#!/bin/sh
if [ -n "$t" ]; then
    echo "Starting server"
    /app/cmd/server/app 9999 $t
elif [ -n "$s" ]; then
    echo "Starting client"
    /app/cmd/client/app 8888 $s
else
    echo "error"
fi