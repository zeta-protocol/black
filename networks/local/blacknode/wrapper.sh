#!/usr/bin/env sh

BINARY=/bkd/linux/${BINARY:-bkd}
echo "binary: ${BINARY}"
ID=${ID:-0}
LOG=${LOG:-bkd.log}

if ! [ -f "${BINARY}" ]; then
	echo "The binary $(basename "${BINARY}") cannot be found. Please add the binary to the shared folder. Please use the BINARY environment variable if the name of the binary is not 'bkd' E.g.: -e BINARY=bkd_my_test_version"
	exit 1
fi

BINARY_CHECK="$(file "$BINARY" | grep 'ELF 64-bit LSB executable, x86-64')"

if [ -z "${BINARY_CHECK}" ]; then
	echo "Binary needs to be OS linux, ARCH amd64"
	exit 1
fi

export BKDHOME="/bkd/node${ID}/bkd"

if [ -d "$(dirname "${BKDHOME}"/"${LOG}")" ]; then
  "${BINARY}" --home "${BKDHOME}" "$@" | tee "${BKDHOME}/${LOG}"
else
  "${BINARY}" --home "${BKDHOME}" "$@"
fi
