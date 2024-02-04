# creavo/system-data-exporter

Goal of this app is to generate a JSON with system-data without having any dependencies or the need to be installed (standalone-app).

It features:

* cpu (type, speed, ...)
* disk (partitions, size, usage, ...)
* host (hostname, operating-system, ...)
* memory (size, usage)
* network (interfaces with addresses)

See [example_linux_amd64.json](https://github.com/creavo/system-data-exporter/blob/main/example_linux_amd64.json) as an example for the output.

### Usage

Just run the app and it will print a json with the data into stdout. If you need, you can add a url as parameter - then the app will send a POST-request to the given url, containing the data in the body. In that case, the app will output the http-code of the response and the content to ease debug.

Example with url: `./system-data-exporter --url=https://example.com/my/endpoint`.

### Todo/Improvements

- [ ] cpu-usage/cpu-load on windows (gopsutil/cpu does not support that)
- [ ] add processes (maybe details per process)

### Alternatives
* Glances (which requires phyton on the system)

This app heavily uses https://github.com/shirou/gopsutil/v3.
