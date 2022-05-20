# Introduction

Test open source libraries in order to wrap them properly.

# Inspect serf api

```bash
## 1. Build the test binary.
make member


## 2. Launch a serf cluster from 3 different terminal.
 ./member -bootstrap=127.0.0.1:8991,127.0.0.1:8992,127.0.0.1:8993 \
     -nodename 8991
     -address=127.0.0.1:8991

 ./member -bootstrap=127.0.0.1:8991,127.0.0.1:8992,127.0.0.1:8993 \
     -nodename 8992 \
     -address=127.0.0.1:8992

 ./member -bootstrap=127.0.0.1:8991,127.0.0.1:8992,127.0.0.1:8993 \
     -nodename 8993 \
     -address=127.0.0.1:8993


## 3. Connect to the started serf cluster from another terminal.
 go test -run=TestCluster -v
```
