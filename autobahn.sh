#!/bin/bash

docker run -it --rm \
  -v "${PWD}/autobahn/config:/config" \
  -v "${PWD}/autobahn/reports:/reports" \
  crossbario/autobahn-testsuite \
  wstest -m fuzzingclient -s /config/fuzzingclient.json