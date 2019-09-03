#!/bin/bash

if [ -n "$MONGO_HOST" ]; then
    sed -i "s/\"mongoHost\" : \"[^\"]*\"/\"mongoHost\" : \"$MONGO_HOST\"/" /app/settings.json
fi

if [ -n "$MONGO_PORT" ]; then 
    sed -i "s/\"mongoPort\" : 27017/\"mongoPort\" : $MONGO_PORT/" /app/settings.json
fi
"$@"
