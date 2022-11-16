#!/usr/bin/env bash

declare -a RPCS=$(echo $RPC_ENDPOINTS | awk -vFS='[ /]+' -vRS='[, ]+'  '{print $2}')

# loop 60 times with a 5 second sleep, 300 seconds total
for i in {1..60}; do
        for RPC in ${RPCS}; do
                echo "Checking for listener at ${RPC}"
                # if there is no port defined skip check
                if [[ "${RPC}" =~ ":" ]]; then
                        timeout 1 bash -c "echo >/dev/tcp/${RPC%:*}/${RPC#*:}" && break 2
                        echo "Connection to ${RPC} failed"
                else
                        break 2
                fi
        done
        echo "No RPC endpoints listening, sleeping 5 seconds"
        sleep 6
done

exec $@
