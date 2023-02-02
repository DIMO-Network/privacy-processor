# Privacy Processor

Allow DIMO users to specify privacy zones, in the form of [H3 indices](https://h3geo.org/docs/highlights/indexing), where all location data will be obscured.

## Documentation

[DIMO documentation](https://docs.dimo.zone/docs)

## Deploy Locally

1. From the root, copy contents of `settings.sample.yaml` using:
   ```
   cp settings.sample.yaml settings.yaml
   ```

2. From the same folder, run processor:
   ```sh
   go run ./cmd/privacy-processor
   ```

## Testing

```
go test ./internal/processors
```

## License

[Apache 2.0](LICENSE)
