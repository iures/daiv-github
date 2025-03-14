# Daiv GitHub

A GitHub integration plugin for the daiv CLI tool. This plugin allows you to generate activity reports from GitHub pull requests, reviews, and comments for use in standup meetings and other contexts.

## Features

- Retrieves GitHub pull requests based on configurable query parameters
- Filters pull requests by time range, base branch, and more
- Intelligently filters out pull requests with no relevant activity in the specified time range
- Supports multiple output formats (JSON, Markdown, HTML)
- Fully configurable queries
- Concurrent processing for improved performance

## Project Structure

- **main.go**: Plugin entry point that exports the Plugin interface
- **plugin/plugin.go**: Core plugin implementation (configuration, lifecycle, etc.)
- **plugin/github/**: Directory containing GitHub integration components
  - **plugin/github/client.go**: GitHub API client implementation
  - **plugin/github/models.go**: Domain models for GitHub data
  - **plugin/github/repository.go**: Data access layer for GitHub
  - **plugin/github/service.go**: Business logic for processing GitHub data
  - **plugin/github/formatters.go**: Output formatters (JSON, Markdown, HTML)
- **Makefile**: Build automation for the plugin

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
   make install
   ```
   
   Or manually:
   ```
   go build -o out/daiv-github.so -buildmode=plugin
   daiv plugin install ./out/daiv-github.so
   ```

## Configuration

This plugin requires the following configuration:

### Required Settings

- **github.username**: Your GitHub username
- **github.organization**: The GitHub organization to monitor
- **github.repositories**: List of repositories to monitor (comma-separated)

### Optional Settings

- **github.format**: Output format (json, markdown, or html)
- **github.query.base_branch**: The base branch to filter pull requests by (default: master)
- **github.query.include_authored**: Whether to include authored pull requests (true/false)
- **github.query.include_reviewed**: Whether to include reviewed pull requests (true/false)

You can configure these settings when you first run daiv after installing the plugin, or by using the `daiv config set` command.

## Usage

After installation and configuration, the plugin will be automatically loaded when you start daiv.

### Generating a Standup Report

```
daiv standup
```

This will generate a report of your GitHub activity for the default time range (usually the last 24 hours). The plugin will automatically filter out pull requests that don't have any comments or changes within the specified time range, ensuring that only relevant activity is included in your report.

### Customizing the Time Range

```
daiv standup --from "2023-03-01" --to "2023-03-14"
```

### Changing the Output Format

You can change the default output format in the configuration, or specify it for a single command:

```
daiv config set github.format html
```

## Architecture

The plugin follows a clean architecture approach with clear separation of concerns:

1. **Domain Models**: Independent data structures representing GitHub entities
2. **Repository Layer**: Handles data access to the GitHub API
3. **Service Layer**: Contains business logic for processing GitHub data
4. **Formatters**: Transform domain models into different output formats
5. **Plugin Layer**: Integrates with the daiv CLI tool

This architecture makes the plugin flexible, maintainable, and testable.

### Performance Optimizations

The plugin includes several performance optimizations:

1. **Concurrent Processing**: Repositories are processed in parallel using goroutines, significantly improving performance for large result sets.
2. **Smart Concurrency**: The plugin automatically switches between sequential and concurrent processing based on the size of the data to avoid overhead for small datasets.
3. **Efficient Data Structures**: The plugin uses appropriate data structures to minimize memory usage and processing time.
4. **Smart Filtering**: The plugin intelligently filters out pull requests that don't have any relevant activity (comments or changes) within the specified time range, reducing noise in your reports.

