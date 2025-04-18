#!/bin/bash

# Should be run with a cronn job:
# -  */30 * * * * /root/shahboard/client-go/rotate_restart.sh >> /root/shahboard/client-go/cron.log 2>&1
# -  0 0 * * * truncate -s 0 /root/shahboard/client-go/cron.log

APP_DIR="./client-go"
APP_BINARY="./bot-runner"
STATE_DIR="/tmp/bot_processes"
BOTS_PER_PROCESS=100

cd "$APP_DIR" || {
  echo "$(date '+%Y-%m-%d %H:%M:%S') Failed to change directory to $APP_DIR"
  exit 1
}

INDEX_FILE="$STATE_DIR/last_restarted_index.txt"
TOTAL_PROCESSES=$(ls "$STATE_DIR"/*.pid 2>/dev/null | wc -l)

if [ "$TOTAL_PROCESSES" -eq 0 ]; then
  echo "$(date '+%Y-%m-%d %H:%M:%S') No bot processes found to rotate."
  exit 1
fi

if [ -f "$INDEX_FILE" ]; then
  LAST_INDEX=$(cat "$INDEX_FILE")
else
  LAST_INDEX=0
fi

RESTART_INDEX=$(( (LAST_INDEX + 1) % TOTAL_PROCESSES ))

PID_FILE="$STATE_DIR/bot_$RESTART_INDEX.pid"
LOG_FILE="$STATE_DIR/bot_$RESTART_INDEX.log"

if [ -f "$PID_FILE" ]; then
  PID=$(cat "$PID_FILE")
  if ps -p "$PID" > /dev/null 2>&1; then
    echo "$(date '+%Y-%m-%d %H:%M:%S') Killing process $PID (bot_$RESTART_INDEX)"
    kill "$PID"
    sleep 2
  else
    echo "$(date '+%Y-%m-%d %H:%M:%S') Process $PID already stopped"
  fi
else
  echo "$(date '+%Y-%m-%d %H:%M:%S') PID file $PID_FILE not found"
fi

FIRST=$((RESTART_INDEX * BOTS_PER_PROCESS))
LAST=$((FIRST + BOTS_PER_PROCESS))

echo "$(date '+%Y-%m-%d %H:%M:%S') Sleeping for 60 seconds before restarting..."
sleep 60

echo "$(date '+%Y-%m-%d %H:%M:%S') Restarting bot_$RESTART_INDEX ($FIRST - $LAST)"
"$APP_BINARY" --firstBot=$FIRST --lastBot=$LAST > "$LOG_FILE" 2>&1 &
echo $! > "$PID_FILE"

echo "$RESTART_INDEX" > "$INDEX_FILE"

echo "$(date '+%Y-%m-%d %H:%M:%S') Done rotating bot_$RESTART_INDEX"
echo "  " 