#!/bin/bash

# first build the binary by `go build -o bot-runner`
# then run this script with the range of bots to start `./start_bots.sh 0 500`
# to stop the bots, run `./stop_all.sh`
# to track the logs, run `tail -f /tmp/bot_processes/*.log`

APP_BINARY="./bot-runner"
STATE_DIR="/tmp/bot_processes"
BOTS_PER_PROCESS=100

FROM=${1:-0}
TO=${2:-400}

mkdir -p "$STATE_DIR"

NUM_PROCESSES=$(((TO - FROM) / BOTS_PER_PROCESS))
for ((i = 0; i < NUM_PROCESSES; i++)); do
  INDEX=$(((FROM / BOTS_PER_PROCESS) + i))
  FIRST=$((INDEX * BOTS_PER_PROCESS))
  LAST=$((FIRST + BOTS_PER_PROCESS))
  LOG_FILE="$STATE_DIR/bot_$INDEX.log"

  echo "Starting bots $FIRST to $LAST..."
  $APP_BINARY --firstBot=$FIRST --lastBot=$LAST > "$LOG_FILE" 2>&1 &
  echo $! > "$STATE_DIR/bot_$INDEX.pid"
done
