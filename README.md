# Privacy Processor

Allow DIMO users to specify privacy zones where all location data will be obscured.

## Documentation

[Dimo documentation](https://docs.dimo.zone/docs)

## Deploy Locally

1. From the root, copy contents of `settings.sample.yaml` using:

```
cp settings.sample.yaml settings.yaml
```

2. From the same folder, run processor:

```
go run ./cmd/privacy-processor
```

## Testing

```
go test ./internal/processors
```

## License

[BSL](LICENSE.md)
