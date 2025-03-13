# Daiv Github

A plugin for the daiv CLI tool.

## Installation

### From GitHub

```
daiv plugin install iures/daiv-github
```

### From Source

1. Clone the repository:
   ```
   git clone https://github.com/iures/daiv-github.git
   cd daiv-github
   ```

2. Build the plugin:
   ```
   go build -o out/daiv-github.so -buildmode=plugin
   ```

3. Install the plugin:
   ```
   daiv plugin install ./out/daiv-github.so
   ```

## Configuration

This plugin requires the following configuration:

- daiv-github.apikey: API key for the service

You can configure these settings when you first run daiv after installing the plugin.

## Usage

After installation, the plugin will be automatically loaded when you start daiv.

