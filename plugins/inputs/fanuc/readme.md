# Telegraf Fanuc Input Plugin

The `fanuc` input plugin gathers metrics from Fanuc CNC machines using the Focas2 library.

## Global configuration options

In addition to the plugin-specific configuration settings, plugins support
additional global and plugin configuration options. These options are described
in the [configuration file documentation][config-file].

## Configuration

```toml @sample.conf
[[inputs.fanuc]]
    ## List of machines to poll for data. The plugin for now assumes port 8192
    machines = ["192.168.0.1", "192.168.0.2"]
    ## Timeout for the machine to respond within which the plugin will skip the machine
    timeout = 10
```

Running `telegraf --usage <plugin-name>` also gives the sample TOML
configuration.
