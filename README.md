# 🎨 pier
Palette SSO kubeconfig CLI

# Installation

Have go installed, then run

```
go install github.com/buzzsurfr/pier@latest
```

## Getting started

You need a credential to login to Palette as a user. The following login options are available:

### API Key

Create an API key from the Palette console and store it in the `PALETTE_API_KEY` environment variable.

### Browser Token

Using **kooky**, pier can access the browser tokens from most browser/OS combinations. Note that on certain platforms may require elevated permissions.

## Commands

### `generate`

Generate (pier gen) will output a kubeconfig file based on the available Palette clusters. These will be named according to the project and cluster name in Palette.

By default, this will output the kubeconfig to stdout, which can be piped into a file.

## Environment Variables

| Key | Description | Default value |
| --- | ----------- | ------------- |
| `PALETTE_HOST` | Palette API endpoint | `api.spectrocloud.com` |
| `PALETTE_API_KEY` | Palette API key | |
| `PALETTE_TOKEN` | JWT from browser | |