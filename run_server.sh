#!/usr/bin/env bash
# Script to run server/main.go periodically via cron.
# Kills any previously running instance before starting a new one.

set -euo pipefail

PROJECT_DIR="/workspace/All-On-Cloud-9"
PIDFILE="$PROJECT_DIR/server.pid"
LOGFILE="$PROJECT_DIR/server.log"

# Kill previous instance if still running
if [ -f "$PIDFILE" ]; then
    OLD_PID=$(cat "$PIDFILE")
    if kill -0 "$OLD_PID" 2>/dev/null; then
        echo "$(date): Stopping previous server (PID $OLD_PID)" >> "$LOGFILE"
        kill "$OLD_PID"
        sleep 2
        # Force kill if still alive
        kill -0 "$OLD_PID" 2>/dev/null && kill -9 "$OLD_PID"
    fi
    rm -f "$PIDFILE"
fi

echo "$(date): Starting server/main.go" >> "$LOGFILE"
cd "$PROJECT_DIR"
go run server/main.go >> "$LOGFILE" 2>&1 &
echo $! > "$PIDFILE"
echo "$(date): Server started (PID $(cat "$PIDFILE"))" >> "$LOGFILE"
