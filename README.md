
Statful Client for golang
==============

Statful client for golang written in go. This client is intended to gather metrics and send them to Statful.

## Table of Contents

* [Supported Versions of Golang](#supported-versions-of-golang)
* [Quick Start](#quick-start)
* [Reference](#reference)
  * [Global Configuration](#global-configuration)
  * [Methods](#methods)
  * [Plugins](#plugins)
* [Examples](#examples)
  * [UDP Configuration](#udp-configuration)
  * [HTTP Configuration](#http-configuration)
  * [Logger Configuration](#logger-configuration)
   * [Configuration of Defaults per Method](#configuration-of-defaults-per-method)
  * [Preset Configuration](#preset-configuration)
  * [Send Metrics Configuration](#send-metrics-configuration)
 * [Authors](#authors)
* [License](#license)

## Supported Versions of Golang

| Statful Client version | Tested golang versions  |
|:---|:---|
| 0.13.0 | `1.15` |
| 0.0.1 | `1.13`, `1.14` |

## Quick Start

Two examples are available in the examples folder. One of an http server and one of a simple metric.
You can use them as guideline on how to setup a golang client on your own project.

## Reference

The following section presents a detailed reference of the available options to take full advantage of Statful.

### Global Configuration

Below you can find the information on the custom options to set up the configurations parameters.

| Option | Description | Type | Default | Required |
|:---|:---|:---|:---|:---|
| _DisableAutoFlush_ | Defines if metrics should be flushed synchronously. ``FlushSize`` and ``FlushInterval`` attributes are disabled and ``Flush()`` or ``FlushError()`` functions should be called instead. | `boolean` | `false` | **NO** |
| _DryRun_ | Defines if metrics should be output to the logger instead of being sent to Statful (useful for testing/debugging purposes). | `boolean` | `false` | **NO** |
| _FlushSize_ | Defines the maximum buffer size before performing a flush, in **bytes**. | `number` | `1000` | **NO** |
| _Globaltags_ | Object for setting the global tags. | `object` | `{}` | **NO** |
| _Url_ | Defines the url where the metrics are sent. | `string` | **none** | **NO** |
| _BasePath_ | Defines the API path to where the metrics are sent. It can only be set inside _api_. | `string` | `/tel/v2.0/metric` | **NO** |
| _token_ | Defines the token used to match incoming data to Statful. It can only be set inside _api_. | `string` | **none** | **YES** |
| _timeout_ | Defines the timeout for the transport layers in **milliseconds**. It can only be set inside _api_. | `number` | `2000` | **NO** |

### Methods

```golang
// Non-Aggregated Metrics
- statful.Counter("myCounter", 1.0);
- statful.CounterWithTags("myCounter", 1, Tags{"foo "bar"});
- statful.Gauge("myGauge", 1.0);
- statful.GaugeWithTags("myGauge", 1, Tags{"foo "bar"});
- statful.Histogram("myHistogram", 1.0);
- statful.HistogramWithTags("myHistogram", 1.0, Tags{"foo "bar"});
- statful.Put(&Metric{Name: "myCustomMetric", Value: 200, Tags{"foo "bar"}, Aggs: Aggregations{AggAvg: struct{}{}}, Freq: Freq30s);
```

The methods for non-aggregated metrics receive a metric name and value as arguments and send a counter, a gauge, a timer or a custom metric.

## Examples

Here you can find some useful usage examples of the Statful’s golang Client.

### UDP Configuration

Create a simple UDP configuration for the client.

```golang
statful.Statful{
    Sender: &statful.ProxyMetricsSender{
        Client: &statful.UdpClient{
            Address: "localhost:1234",
        },
    },
    GlobalTags: statful.Tags{"client": "golang"},
    DryRun:     false,
}
```

### HTTP Configuration

Create a simple HTTP API configuration for the client.

```golang
statful.New(
    statful.Configuration{
        DryRun: false,
        FlushSize: 1000,

        Sender: &statful.HttpSender{
            Http:     &http.Client{},
            Url:      "https://api.statful.com/tel/v2.0/",
            Token:    "12345678-09ab-cdef-1234-567890abcdef",
        },

        Logger: log.New(os.Stderr, "", log.LstdFlags),
        Tags: statful.Tags{"client": "golang"},
    }
)
```

### Disabling Auto Flush

Create an HTTP API configuration that prevents flushing asynchronously.

```golang
statful.New(
    statful.Configuration{
        DisableAutoFlush: true,
        DryRun: false,

        Sender: &statful.HttpSender{
            Http:     &http.Client{},
            Url:      "https://api.statful.com/tel/v2.0/",
            Token:    "12345678-09ab-cdef-1234-567890abcdef",
        },

        Logger: log.New(os.Stderr, "", log.LstdFlags),
        Tags: statful.Tags{"client": "golang"},
    }
)
```

### Buffer Configuration

Create a simple Metrics Sender that buffers metrics before sending.

```golang
statful.New(
    statful.Configuration{
        DryRun: false,
        FlushSize: 1000,

        Sender: &statful.HttpSender{
            Http:     &http.Client{},
            Url:      "https://api.statful.com/tel/v2.0/",
            Token:    "12345678-09ab-cdef-1234-567890abcdef",
        },

        Logger: log.New(os.Stderr, "", log.LstdFlags),
        Tags: statful.Tags{"client": "golang"},
    }
)
```


### Event Sender Configuration

To send event payload (JSON) you may configure event sender in your client.

```golang
statful.New(
    statful.Configuration{
        DryRun: false,

        EventSender: &statful.HttpSender{
            Http:     &http.Client{},
            Url:      "https://api.statful.com/insights/event/",
            Token:    "12345678-09ab-cdef-1234-567890abcdef",
        },
        EventLogger: log.New(os.Stderr, "", log.LstdFlags),
    }
)
```

## Authors

[Statful](https://github.com/Statful)

## License

Statful Golang Client is available under the MIT license. See the [LICENSE](https://raw.githubusercontent.com/statful/statful-client-objc/master/LICENSE) file for more information.

