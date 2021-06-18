package haproxy

import (
	"reflect"
	"testing"
)

func TestConfigParser(t *testing.T) {
	const (
		testHAProxyConfigPath = "./data/haproxy.config"
		testNs                = "test-route"
		testFeName            = "public"
		notExistingBe         = "notexistingtestbackend"
		emptyBe               = "empty"
	)

	var (
		expectedDefaults = []string{
			"maxconn 20000",
			"option httplog",
			"log global",
			"errorfile 503 /var/lib/haproxy/conf/error-page-503.http",
			"errorfile 404 /var/lib/haproxy/conf/error-page-404.http",
			"timeout connect 5s",
			"timeout client 30s",
			"timeout client-fin 1s",
			"timeout server 30s",
			"timeout server-fin 1s",
			"timeout http-request 10s",
			"timeout http-keep-alive 300s",
			"timeout tunnel 1h",
		}
		expectedGlobal = []string{
			"maxconn 20000",
			"nbthread 4",
			"daemon",
			"log /var/lib/rsyslog/rsyslog.sock local1 debug",
			"log-send-hostname",
			"ca-base /etc/ssl",
			"crt-base /etc/ssl",
			"stats socket /var/lib/haproxy/run/haproxy.sock mode 600 level admin expose-fd listeners",
			"stats timeout 2m",
			"tune.maxrewrite 8192",
			"tune.bufsize 32768",
			"ssl-default-bind-options ssl-min-ver TLSv1.2",
			"tune.ssl.default-dh-param 2048",
			"ssl-default-bind-ciphers TLS_AES_128_GCM_SHA256:TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384",
		}
		expectedBackends = map[string][]string{
			"be_http:test-route:hello-openshift": []string{
				"mode http",
				"option redispatch",
				"option forwardfor",
				"balance",
				"timeout check 5000ms",
				"http-request add-header X-Forwarded-Host %[req.hdr(host)]",
				"http-request add-header X-Forwarded-Port %[dst_port]",
				"http-request add-header X-Forwarded-Proto http if !{ ssl_fc }",
				"http-request add-header X-Forwarded-Proto https if { ssl_fc }",
				"http-request add-header X-Forwarded-Proto-Version h2 if { ssl_fc_alpn -i h2 }",
				"http-request add-header Forwarded for=%[src];host=%[req.hdr(host)];proto=%[req.hdr(X-Forwarded-Proto)]",
				"cookie 2659ff115eb04a4de61d24b69643bf51 insert indirect nocache httponly",
				"server pod:hello-openshift-7b8c68587c-mtck6:hello-openshift:8080-tcp:10.217.0.24:8080 10.217.0.24:8080 cookie 97ddca1460eca46376928fd7bcf8c89c weight 256 check inter 5000ms",
			},
			"be_edge_http:test-route:hello-openshift2": []string{
				"mode http",
				"option redispatch",
				"option forwardfor",
				"balance",
				"timeout check 5000ms",
				"http-request add-header X-Forwarded-Host %[req.hdr(host)]",
				"http-request add-header X-Forwarded-Port %[dst_port]",
				"http-request add-header X-Forwarded-Proto http if !{ ssl_fc }",
				"http-request add-header X-Forwarded-Proto https if { ssl_fc }",
				"http-request add-header X-Forwarded-Proto-Version h2 if { ssl_fc_alpn -i h2 }",
				"http-request add-header Forwarded for=%[src];host=%[req.hdr(host)];proto=%[req.hdr(X-Forwarded-Proto)]",
				"cookie f7e936ebe97cd64e8c888d1eba08cb29 insert indirect nocache httponly secure attr SameSite=None",
				"server pod:hello-openshift-7b8c68587c-mtck6:hello-openshift:8080-tcp:10.217.0.24:8080 10.217.0.24:8080 cookie 97ddca1460eca46376928fd7bcf8c89c weight 256 check inter 5000ms",
			},
			"be_http:test-route:httpd": []string{
				"mode http",
				"option redispatch",
				"option forwardfor",
				"balance",
				"timeout check 5000ms",
				"http-request add-header X-Forwarded-Host %[req.hdr(host)]",
				"http-request add-header X-Forwarded-Port %[dst_port]",
				"http-request add-header X-Forwarded-Proto http if !{ ssl_fc }",
				"http-request add-header X-Forwarded-Proto https if { ssl_fc }",
				"http-request add-header X-Forwarded-Proto-Version h2 if { ssl_fc_alpn -i h2 }",
				"http-request add-header Forwarded for=%[src];host=%[req.hdr(host)];proto=%[req.hdr(X-Forwarded-Proto)]",
				"cookie 300252a1790569894d23351f1f069d83 insert indirect nocache httponly",
				"server pod:httpd-7c7ccfffdc-kxg8v:httpd:8080-tcp:10.217.0.22:8080 10.217.0.22:8080 cookie d21b524b661cd1d9d750a522b3dd7edc weight 256 check inter 5000ms",
			},
			"be_edge_http:test-route:httpd2": []string{
				"mode http",
				"option redispatch",
				"option forwardfor",
				"balance",
				"acl whitelist src 2600:14a0::/40",
				"tcp-request content reject if !whitelist",
				"timeout check 5000ms",
				"http-request add-header X-Forwarded-Host %[req.hdr(host)]",
				"http-request add-header X-Forwarded-Port %[dst_port]",
				"http-request add-header X-Forwarded-Proto http if !{ ssl_fc }",
				"http-request add-header X-Forwarded-Proto https if { ssl_fc }",
				"http-request add-header X-Forwarded-Proto-Version h2 if { ssl_fc_alpn -i h2 }",
				"http-request add-header Forwarded for=%[src];host=%[req.hdr(host)];proto=%[req.hdr(X-Forwarded-Proto)]",
				"cookie 1c6f5d1acb56fe6e379eaf39b37d10ef insert indirect nocache httponly secure attr SameSite=None",
				"server pod:httpd-7c7ccfffdc-kxg8v:httpd:8080-tcp:10.217.0.22:8080 10.217.0.22:8080 cookie d21b524b661cd1d9d750a522b3dd7edc weight 256 check inter 5000ms",
			},
		}
		expectedFrontends = map[string][]string{
			"public": []string{
				"bind :80",
				"mode http",
				"tcp-request inspect-delay 5s",
				"tcp-request content accept if HTTP",
				"monitor-uri /_______internal_router_healthz",
				"http-request del-header Proxy",
				"http-request set-header Host %[req.hdr(Host),lower]",
				"acl secure_redirect base,map_reg(/var/lib/haproxy/conf/os_route_http_redirect.map) -m found",
				"redirect scheme https if secure_redirect",
				"use_backend %[base,map_reg(/var/lib/haproxy/conf/os_http_be.map)]",
				"default_backend openshift_default",
			},
			"public_ssl": []string{
				"option tcplog",
				"bind :443",
				"tcp-request  inspect-delay 5s",
				"tcp-request content accept if { req_ssl_hello_type 1 }",
				"acl sni req.ssl_sni -m found",
				"acl sni_passthrough req.ssl_sni,lower,map_reg(/var/lib/haproxy/conf/os_sni_passthrough.map) -m found",
				"use_backend %[req.ssl_sni,lower,map_reg(/var/lib/haproxy/conf/os_tcp_be.map)] if sni sni_passthrough",
				"use_backend be_sni if sni",
				"default_backend be_no_sni",
			},
		}
	)

	parser := NewConfigParser(testHAProxyConfigPath)
	if err := parser.Parse(); err != nil {
		t.Fatalf("Failed to parse the config:%s", err)
	}

	t.Log("Getting global section")
	if !reflect.DeepEqual(expectedGlobal, parser.GlobalSection) {
		t.Errorf("Expected global section %v,\ngot%v", expectedGlobal, parser.GlobalSection)
	}

	t.Log("Getting defaults section")
	if !reflect.DeepEqual(expectedDefaults, parser.DefaultsSection) {
		t.Errorf("Expected defaults section %v,\ngot%v", expectedDefaults, parser.DefaultsSection)
	}

	t.Logf("Getting multiple backends")
	gotBeMap := parser.Backends(testNs)
	if len(gotBeMap) != len(expectedBackends) {
		t.Errorf("Expected %d backends for namespace %s, got %d", len(expectedBackends), testNs, len(gotBeMap))
	}

	t.Log("Checking backend names and contents")
	if !reflect.DeepEqual(gotBeMap, expectedBackends) {
		t.Errorf("Expected %v backends,\ngot %v", expectedBackends, gotBeMap)
	}

	t.Log("Getting every backend by name")
	for name, contents := range expectedBackends {
		gotContents, _ := parser.Backend(name)
		if !reflect.DeepEqual(gotContents, contents) {
			t.Errorf("Backend %s expected to have %v\nbut got %v", name, contents, gotContents)
		}
	}

	t.Logf("Getting multiple frontends")
	gotFeMap := parser.Frontends(testFeName)
	if len(gotBeMap) != len(expectedBackends) {
		t.Errorf("Expected %d backends for name substring %s, got %d", len(expectedFrontends), testFeName, len(gotFeMap))
	}

	t.Log("Checking frontend names and contents")
	if !reflect.DeepEqual(gotFeMap, expectedFrontends) {
		t.Errorf("Expected %v frontends,\ngot %v", expectedFrontends, gotFeMap)
	}

	t.Log("Getting every frontend by name")
	for name, contents := range expectedFrontends {
		gotContents, _ := parser.Frontend(name)
		if !reflect.DeepEqual(gotContents, contents) {
			t.Errorf("Frontend %s expected to have %v\nbut got %v", name, contents, gotContents)
		}
	}

	t.Log("Not existing backend")
	gotBeMap = parser.Backends(notExistingBe)
	if len(gotBeMap) > 0 {
		t.Errorf("Expected no backend for %s,\ngot %v", notExistingBe, gotBeMap)
	}
	_, exists := parser.Backend(notExistingBe)
	if exists {
		t.Errorf("Backend %s not expected to exist", notExistingBe)
	}

	t.Log("Empty backend")
	_, exists = parser.Backend(emptyBe)
	if !exists {
		t.Errorf("Backend %s expected to exist", emptyBe)
	}
}
