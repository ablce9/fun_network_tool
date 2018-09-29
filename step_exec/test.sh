#! /bin/bash

echo 'sleep 1
sleep 1
sleep 1' | go run step_exec.go -cmd '/bin/sh -c' -step 3
