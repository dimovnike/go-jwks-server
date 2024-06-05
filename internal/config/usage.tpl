Usage:

Flags can be provided via environment variables by prefixing the flag name with {{.envVarPrefix}}, replacing dashes with underscore and converting it to uppercase. Example: flag -{{.exampleFlag}} can be provided via environment variable {{.exampleEnvVar}}.

NOTE: It is not allowed to provide a flag both in the command line and in the environment variable.

The -key-dir directory must contain the public keys, one key in a file. The file name is the key ID, files my have an optional .pub extension.  Files that have .ignore extension are ignored.

Supported flags:
{{/* keep this line last */}}