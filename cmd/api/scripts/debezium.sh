#!/bin/bash
set -e

CONNECT_URL="${CONNECT_URL:-http://localhost:8083}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CONNECTOR_DIR="$SCRIPT_DIR/../debezium"

usage() {
    echo "Usage: $0 <command> [connector-name]"
    echo ""
    echo "Commands:"
    echo "  list              List all connectors"
    echo "  status [name]     Show connector status (all if no name)"
    echo "  create-all        Create all connectors from docker/debezium/*.json"
    echo ""
    echo "Environment variables:"
    echo "  CONNECT_URL       Kafka Connect REST API URL (default: http://localhost:8083)"
    exit 1
}

wait_for_connect() {
    echo "Waiting for Kafka Connect to be ready..."
    until curl -s "$CONNECT_URL/connectors" > /dev/null 2>&1; do
        sleep 2
    done
    echo "Kafka Connect is ready"
}

case "${1:-}" in
    list)
        curl -s "$CONNECT_URL/connectors" | jq .
        ;;
    status)
        if [ -n "${2:-}" ]; then
            curl -s "$CONNECT_URL/connectors/$2/status" | jq .
        else
            for conn in $(curl -s "$CONNECT_URL/connectors" | jq -r '.[]'); do
                echo "=== $conn ==="
                curl -s "$CONNECT_URL/connectors/$conn/status" | jq .
            done
        fi
        ;;
    create-all)
        wait_for_connect
        for file in "$CONNECTOR_DIR"/connector-*.json; do
            [ -f "$file" ] || continue
            name=$(jq -r '.name' "$file")
            echo "Creating connector: $name"
            curl -s -X POST "$CONNECT_URL/connectors" \
                -H "Content-Type: application/json" \
                -d @"$file" | jq .
        done
        ;;
    *)
        usage
        ;;
esac
