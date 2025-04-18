#!/bin/bash

STATE_DIR="/tmp/bot_processes"

if [ ! -d "$STATE_DIR" ]; then
  echo "No bot state directory found at $STATE_DIR"
  exit 1
fi

echo "Stopping all bot processes..."

for pidfile in "$STATE_DIR"/*.pid; do
  if [ -f "$pidfile" ]; then
    pid=$(cat "$pidfile")
    if ps -p "$pid" > /dev/null 2>&1; then
      echo "Stopping PID $pid from $pidfile"
      kill "$pid"
    else
      echo "PID $pid not running (from $pidfile)"
    fi
    rm -f "$pidfile"
  fi
done

echo "All bot processes stopped."
