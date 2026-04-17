#!/bin/bash

if [[ $1 == "build" ]]; then 
    echo "building go app"
    make build
elif [[ $1 == "run" ]]; then 
    make run 
fi

