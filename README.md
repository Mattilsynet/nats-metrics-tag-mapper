# NATS Account Mapper

This tool fetches all NATS Account IDs from the `/accountz` endpoint of a NATS server (default: `http://localhost:8222`).  
It then retrieves the corresponding account names and generates a Starlark script that maps account IDs to account names.  
The output file is written to `/etc/telegraf/add_account_name.star` by default, for use with Telegraf metric processing.

## Getting started

1. **Build the binary:**
   ```sh
   go build -o nats-metrics-tag-mapper ./cmd/nats-metrics-tag-mapper
   ```

1. **Run the tool:**
   ```sh
   ./nats-metrics-tag-mapper
   ```
   - Use `-output` to specify a custom output file path.
   - Use `-url` to specify a custom NATS metrics endpoint.

## Example

```sh
./nats-metrics-tag-mapper -output /tmp/add_account_name.star -url http://nats-server:8222
```

## Output

The generated Starlark file defines an `apply(metric)` function that adds an `account_name` tag to metrics based on the account ID.
For more details about Telegraf usage see [Telegraf at Github](https://github.com/influxdata/telegraf)

## License

See [`LICENSE`](LICENSE) for details.
