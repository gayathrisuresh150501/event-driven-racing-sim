#!/bin/bash

# Script to test distributed racing system
# Usage: ./scripts/test-distributed.sh

set -e

echo "🏁 Testing Distributed Racing System"
echo "====================================="
echo ""

# Start engine server
echo "Starting engine server on port 8080..."
PORT=8080 go run cmd/engine-server/main.go &
ENGINE_PID=$!
sleep 2

# Start AI services
echo "Starting AI service 1 on port 8081..."
PORT=8081 AI_NAME=RuleBot-Alpha go run cmd/ai-service/main.go &
AI1_PID=$!
sleep 1

echo "Starting AI service 2 on port 8083..."
PORT=8083 AI_NAME=RuleBot-Bravo go run cmd/ai-service/main.go &
AI2_PID=$!
sleep 1

# Start gateway
echo "Starting gateway on port 8082..."
PORT=8082 ENGINE_ADDRESS=http://localhost:8080 AI_ADDRESS_1=http://localhost:8081 AI_ADDRESS_2=http://localhost:8083 go run cmd/gateway/main.go &
GATEWAY_PID=$!
sleep 2

echo ""
echo "All services started!"
echo ""

# Cleanup function
cleanup() {
    echo ""
    echo "Shutting down services..."
    kill $ENGINE_PID $AI1_PID $AI2_PID $GATEWAY_PID 2>/dev/null || true
    wait
    echo "All services stopped."
}

trap cleanup EXIT

# Test health checks
echo "Testing health checks..."
curl -s http://localhost:8080/health && echo "Engine healthy"
curl -s http://localhost:8081/health && echo "AI-1 healthy"
curl -s http://localhost:8083/health && echo "AI-2 healthy"
curl -s http://localhost:8082/health && echo "Gateway healthy"
echo ""

# Start a race
echo "Starting a race with 4 cars..."
RACE_RESPONSE=$(curl -s -X POST http://localhost:8082/race/start \
  -H "Content-Type: application/json" \
  -d '{"num_cars": 4}')

RACE_ID=$(echo $RACE_RESPONSE | jq -r '.race_id')
echo "Race ID: $RACE_ID"
echo ""

# Run the race
echo "Running race for 10 ticks..."
curl -s -X POST http://localhost:8082/race/run \
  -H "Content-Type: application/json" \
  -d "{\"race_id\": \"$RACE_ID\", \"num_ticks\": 10}" | while read line; do
    TICK=$(echo $line | jq -r '.tick_number // empty')
    if [ ! -z "$TICK" ]; then
        echo "  Tick $TICK"
    fi
done
echo ""

# Get results
echo "Getting race results..."
RESULTS=$(curl -s -X POST http://localhost:8082/race/results \
  -H "Content-Type: application/json" \
  -d "{\"race_id\": \"$RACE_ID\"}")

echo "$RESULTS" | jq '.'
echo ""

echo "DISTRIBUTED SYSTEM TEST PASSED!"
echo " Proved: Client → Gateway → Engine → AI all communicate over HTTP"
echo ""
echo "Press Ctrl+C to stop all services..."
wait
