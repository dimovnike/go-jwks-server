Server for JSON Web Key sets (JWKS).

Designed to play nicely with kubernetes and other orchestration systems (like docker swarm).

## Main features:

- Serve JWKS from a directory with public PEM files. File names are used as key IDs.
- Can watch the directory for changes and reload the keys (useful with kubernetes secrets).
- Sets cache control headers according to the config.
- Can be configured using command line flags and environment variables.
- Configurable logging using zerolog.

## Example usage:

Create the kubernetes pod (service, ingress, etc):

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: go-jwks-server
  labels:
    app: go-jwks-server
spec:
  containers:
    - name: go-jwks-server
      image: ghcr.io/dimovnike/go-jwks-server:latest
      ports:
        - containerPort: 8080
      volumeMounts:
        - name: jwks-keys
          mountPath: /jwt-keys
      env:
        - name: GO_JWKS_SERVER_KEY_DIR
          value: /jwt-keys
        - name: GO_JWKS_SERVER_LOG_LEVEL
          value: debug
  volumes:
    - name: jwks-keys
      secret:
        secretName: jwks-public-keys
        optional: true

```

NOTE: the secret does not yet exist.

Now try accessing the service, it will produce empty keys as expected:

```sh
curl -s localhost:8080/keys | jq
{
  "keys": []
}
```

Create the secret with the public keys:

```bash
# generate a key pair
openssl ec -in priv-key.pem -pubout > pub-key.pem
openssl ec -pubin -in pub-key.pem -text

# create and apply the secret
kubectl create secret generic jwks-public-keys --from-file=key1=pub-key.pem  --dry-run=client --save-config --dry-run=client -oyaml \
    | kubectl apply -f -

```

NOTE: you can run the last command multiple times to add more keys to the secret.

Wait for a while for the secret to propagate to the pod, you will see in the log:

```
{"level":"info","skipped":{"..2024_06_05_16_49_04.104114561":"directory","..data":"directory"},"loaded":{"key1":"key1"},"time":"2024-06-05T16:49:05Z","caller":"/build/internal/keyloader/keys.go:83","message":"loaded keys"}
```

 and try accessing the service again:

```sh
curl -s localhost:8080/keys | jq
{
  "keys": [
    {
      "crv": "P-256",
      "kid": "key1",
      "kty": "EC",
      "x": "wYthzYx55RuKWa8Ru3Tp2_LcbBixW3TXjEkkt5ZrVsY",
      "y": "N2V-3e35eQrAxCyaXRnnV1IQVnhgVw2TgWZ6UZmj5n0"
    }
  ]
}
```

We can see the public key with the ID `key1` is now present in the JWKS.


## Command line flags

```text
Flags can be provided via environment variables by prefixing the flag name with GO_JWKS_SERVER_, replacing dashes with underscore and converting it to uppercase. Example: flag -dir-watch-interval can be provided via environment variable GO_JWKS_SERVER_DIR_WATCH_INTERVAL.

NOTE: It is not allowed to provide a flag both in the command line and in the environment variable.

The -key-dir directory must contain the public keys, one key in a file. The file name is the key ID, files my have an optional .pub extension.  Files that have .ignore extension are ignored.

Supported flags:

  -dir-watch-interval duration
        the interval to check the key directory for changes, set to 0 to disable watching (default 1s)
  -exit-on-error
        exit if loading keys fails
  -http-addr string
        the address to listen on (default ":8080")
  -http-cache-max-age duration
        set max-age in the cache-control header in seconds, set to 0 to disable caching (default 1h0m0s)
  -http-idle-timeout duration
        the maximum amount of time to wait for the next request when keep-alives are enabled
  -http-keys-endpoint string
        the endpoint to serve the keys (default "/keys")
  -http-max-header-bytes int
        the maximum number of bytes the server will read parsing the request headers, including the request line (default 1048576)
  -http-read-header-timeout duration
        timeout for reading the request headers
  -http-read-timeout duration
        timeout for reading the entire request, including the body
  -http-shutdown-timeout duration
        timeout for graceful shutdown of the server (default 5s)
  -http-write-timeout duration
        timeout for writing the response
  -key-dir string
        the directory to load the keys from (default "./keys")
  -log-caller
        show caller file and line number (default true)
  -log-console
        human friendly logs on console
  -log-file string
        log file, stdout or stderr (default "stderr")
  -log-level string
        logging level (debug, info, warn, error, fatal, panic, no, disabled, trace) (default "error")
  -log-no-color
        do not show color on console
  -log-stack
        show stack info
  -log-timestamp
        show timestamp (default true)
  -print-config
        print the configuration and exit

```