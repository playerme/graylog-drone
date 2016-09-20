package drone

import (
	"fmt"
	"github.com/vjeantet/grok"
)

var g *grok.Grok

func init() {
	g, _ = grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	g.AddPattern("HAPROXY_HTTP", `%{IP:client_ip}:%{POSINT:client_post} \[%{DATA:timestamp}\] %{DATA:frontend_name_transport} %{DATA:backend_name}/%{DATA:server_name} %{NUMBER:queue_time}/%{NUMBER:waiting_time}/%{NUMBER:connect_time}/%{NUMBER:response_time}/%{NUMBER:total_time} %{POSINT:status_code} %{NUMBER:bytes_read} (?:-|%{DATA:captured_request_cookie}) (?:-|%{DATA:captured_response_cookie}) --(?:--|%{DATA:termination_state}) %{NUMBER:process_active_connections}/%{NUMBER:process_frontend_connections}/%{NUMBER:process_backend_connections}/%{NUMBER:process_awaiting_connections}/%{NUMBER:retries} %{NUMBER:server_queue}/%{NUMBER:backend_queue} %{DATA:captured_request_headers} %{DATA:captured_response_headers} "(?:%{WORD:verb} %{URIPATHPARAM:url}(?: HTTP/%{NUMBER:httpversion})?|-)"`)
	g.AddPattern("NGINX_TIME", `%{IPORHOST:clientip} - - \[%{HTTPDATE:timestamp}\] "(?:%{WORD:verb} %{URIPATHPARAM:request}(?: HTTP/%{NUMBER:httpversion})?|-)" %{NUMBER:response} (?:%{NUMBER:bytes}|-) "(?:%{URI:referrer}|-)" "%{DATA:agent}" ?(?:%{NUMBER:request_time}|) ?(?:%{NUMBER:upstream_response_time}|)`)
}

func GrokParser(text string, config LogConfig) (out string, data map[string]interface{}, err error) {

	pattern := config.Pattern

	data, err = g.ParseTyped(pattern, text)
	if err != nil {
		return out, data, err
	}

	switch config.ShortText {
	case "http_request":
		out = fmt.Sprintf("%s %s %s", data["response"], data["verb"], data["request"])
		break
	case "haproxy_request":
		out = fmt.Sprintf("%s %s %s %s/%s", data["status_code"], data["verb"], data["url"], data["backend_name"], data["server_name"])
		break
	default:
		out = text
	}

	return out, data, err
}
