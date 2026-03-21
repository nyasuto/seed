#!/usr/bin/env bash
set -euo pipefail

# Ralph Loop - autonomous Claude Code iteration
# Usage: ./ralph.sh [max_iterations]
#
# Runs Claude Code in a loop, reading PROMPT.md each iteration.
# Claude reads tasks.md, picks the next unchecked task, implements it,
# runs tests, and checks it off. Loop continues until all tasks are done
# or max_iterations is reached.

MAX_ITER="${1:-50}"
LOGDIR="./logs"
mkdir -p "$LOGDIR"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOGFILE="${LOGDIR}/ralph_${TIMESTAMP}.log"

echo "=== Ralph Loop started at $(date) ===" | tee "$LOGFILE"
echo "=== Max iterations: ${MAX_ITER} ===" | tee -a "$LOGFILE"

for i in $(seq 1 "$MAX_ITER"); do
  echo "" | tee -a "$LOGFILE"
  echo "--- Iteration ${i}/${MAX_ITER} at $(date) ---" | tee -a "$LOGFILE"

  # Check if all tasks are done
  if ! grep -q '^\- \[ \]' tasks.md 2>/dev/null; then
    echo "=== All tasks completed! ===" | tee -a "$LOGFILE"
    exit 0
  fi

  # Count remaining tasks
  REMAINING=$(grep -c '^\- \[ \]' tasks.md 2>/dev/null || echo "0")
  echo "Remaining tasks: ${REMAINING}" | tee -a "$LOGFILE"

  # Run Claude Code
  claude --print --dangerously-skip-permissions "$(cat PROMPT.md)" 2>&1 | tee -a "$LOGFILE"

  EXIT_CODE=${PIPESTATUS[0]}
  if [ "$EXIT_CODE" -ne 0 ]; then
    echo "!!! Claude exited with code ${EXIT_CODE}, retrying... !!!" | tee -a "$LOGFILE"
    sleep 5
  fi

  # Brief pause between iterations
  sleep 2
done

echo "=== Ralph Loop finished (max iterations reached) at $(date) ===" | tee -a "$LOGFILE"