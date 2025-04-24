# Snapbuild Development Guide

## Prerequisites

### Installing Go 1.16.15

Snapbuild requires Go version 1.16.15. To install it:

1. Download Go 1.16.15:
```bash
curl -L https://golang.org/dl/go1.16.15.darwin-arm64.tar.gz -o go1.16.15.tar.gz
```

2. Extract and set up:
```bash
mkdir -p ~/go1.16.15
tar -C ~/go1.16.15 -xzf go1.16.15.tar.gz
```

3. Add to your shell configuration (for zsh):
```bash
echo 'export GOROOT=~/go1.16.15/go' >> ~/.zshrc
echo 'export PATH=$GOROOT/bin:$PATH' >> ~/.zshrc
```

4. Reload your shell configuration:
```bash
source ~/.zshrc
```

5. Verify the installation:
```bash
go version  # Should show go1.16.15
```

## Development Workflow

### Testing

To run the Go tests:
```bash
make test-go
```

### Building NPM Dependencies

To build all NPM dependencies locally:
```bash
make snap-platform-all
```

## Release Process

1. Update the version number:
   - Edit `version.snap.txt` with the new version number

2. Build the NPM dependencies:
   - Run `make snap-platform-all` to build all platform-specific packages

3. Commit and push changes:
   - Commit all local changes
   - Issue a PR and merge the changes into the `main` branch

4. Publish to NPM:
   - Log into NPM (credentials available in 1Password)
   - Run `make snap-publish-all` entering the one-time passcode from 1Password when prompted:
     ```bash
     make snap-publish-all
     ```

## Troubleshooting

If you encounter any issues with the Go version check, ensure that:
1. Go 1.16.15 is properly installed
2. The GOROOT and PATH environment variables are correctly set
3. You're using the correct shell configuration file for your shell

For NPM publishing issues:
1. Verify you're logged into NPM
2. Ensure you have the correct one-time passcode
3. Check that all platform-specific packages were built successfully
