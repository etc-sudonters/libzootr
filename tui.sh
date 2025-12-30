#!/usr/bin/env bash

DUMP_ZOOTR_FROM=""
SPOILER_PATH=""
DUMP_DIR=".data"
LOG_DIR=".logs"
LOGIC="glitchless"

function stderr() {
    >&2 echo "${*}"
}

function fatal() {
    EC="${1}"
    shift
    stderr "${0}" "${@}"
    exit "${EC}"
}

while [ "$#" -gt 0 ]; do
    case "${1}" in
        "-Z" | "--zootr" )
            shift
            DUMP_ZOOTR_FROM="${1}"
        ;;
        "-S" | "--spoiler" )
            shift
            SPOILER_PATH="${1}"
        ;;
        "-D" | "--dump-dir" )
            shift
            DUMP_DIR="${1}"
        ;;
        "-G" | "--glitched" )
            shift
            LOGIC="glitched"
        ;;
        "--log-dir" )
            shift
            LOG_DIR="${1}"
        ;;
        * )
            fatal 4 "Unknown flag ${1}"
        ;;
    esac
    shift
done


ZOOTR_DIR="${DUMP_DIR}/zootr"
mkdir -p "${LOG_DIR}"
mkdir -p "${ZOOTR_DIR}"

[ -z "${SPOILER_PATH}" ] && SPOILER_PATH="${DUMP_DIR}/spoiler.json"
[ -n "${DUMP_ZOOTR_FROM}" ] && ./dump-zootr.py --zootr "${DUMP_ZOOTR_FROM}" --output "${ZOOTR_DIR}" | >&2 tee "${LOG_DIR}/zootr-dump"
stderr "launching libzootr tui"
exec go run ./cmd/knowitall/ \
    -debug \
    -logdir "${LOG_DIR}" \
    -world "${ZOOTR_DIR}/logic/${LOGIC}" \
    -data "${ZOOTR_DIR}/data" \
    -spoiler "${SPOILER_PATH}" \
    --
