# gsr - Google Search Relay

A dead simple Google proxy that uses Google Custom Search API under the hood. Can be used to provide service to poor guys who cannot access Google.com.

# How to use

```
go run main.go <conf_file>
```

where in `<conf_file>` you specify your Google Custom Search engine ID and API key:

```
{
  "port": 8080,
  "apiKey": "YOUR_API_KEY_HERE",
  "engineID": "YOUR_ENGINE_ID_HERE",
  "timeout": "5s"
}
```

The website can then be accessed via port `8080`. `timeout` is a timeout for each request to Google Custom Search API.
