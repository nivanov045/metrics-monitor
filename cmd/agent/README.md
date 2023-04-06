# Agent
Collects and sends runtime metrics of the system at a specified frequency. You can add a key for encrypting data using SHA256.
## Agent Settings
* command line flag `a` or environment variable `ADDRESS` to specify the server address, `127.0.0.1:8080` by default
* command line flag `p` or environment variable `POLL_INTERVAL` to specify intervals between metric measurements, 2 seconds by default
* command line flag `r` or environment variable `REPORT_INTERVAL` to specify intervals between sending metrics, 5 seconds by default
* command line flag `k` or environment variable `KEY` to specify the encryption key
