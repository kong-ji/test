package fingerprint

import (
	"path/filepath"
	"testing"

	"github.com/kong-ji/test/internal/rules"
)

func loadEngine(t *testing.T) *Engine {
	t.Helper()
	path := filepath.Join("..", "..", "configs", "rules.json")
	r, err := rules.Load(path)
	if err != nil {
		t.Fatalf("load rules: %v", err)
	}
	return New(r)
}

func TestIdentifyCases(t *testing.T) {
	e := loadEngine(t)

	cases := []struct {
		name string
		in   Input
		want Result
	}{
		{"ssh ubuntu", Input{"1.2.3.4", 22, "SSH-2.0-OpenSSH_8.9p1 Ubuntu-3"},
			Result{"1.2.3.4", 22, "SSH", "OpenSSH", "8.9p1", "Ubuntu", 0.95}},
		{"http nginx", Input{"1.2.3.5", 80, "HTTP/1.1 200 OK\r\nServer: nginx/1.24.0\r\nContent-Type: text/html"},
			Result{"1.2.3.5", 80, "HTTP", "nginx", "1.24.0", "", 0.9}},
		{"http apache", Input{"1.2.3.6", 443, "HTTP/1.1 200 OK\r\nServer: Apache/2.4.57"},
			Result{"1.2.3.6", 443, "HTTP", "Apache", "2.4.57", "", 0.9}},
		{"mysql handshake 8.0.32", Input{"1.2.3.7", 3306, string([]byte{'J', 0, 0, 0, '\n', '8', '.', '0', '.', '3', '2', 0})},
			Result{"1.2.3.7", 3306, "MySQL", "MySQL", "8.0.32", "", 0.9}},
		{"redis err", Input{"1.2.3.8", 6379, "-ERR wrong number of arguments for 'get' command"},
			Result{"1.2.3.8", 6379, "Redis", "Redis", "", "", 0.7}},
		{"ftp proftpd", Input{"1.2.3.9", 21, "220 ProFTPD 1.3.7 Server (ProFTPD)"},
			Result{"1.2.3.9", 21, "FTP", "ProFTPD", "1.3.7", "", 0.9}},
		{"http jetty", Input{"1.2.3.10", 8080, "HTTP/1.1 404 Not Found\r\nServer: Jetty/9.4.51"},
			Result{"1.2.3.10", 8080, "HTTP", "Jetty", "9.4.51", "", 0.85}},
		{"ssh debian", Input{"1.2.3.11", 22, "SSH-2.0-OpenSSH_9.3 Debian-1"},
			Result{"1.2.3.11", 22, "SSH", "OpenSSH", "9.3", "Debian", 0.95}},
		{"http nginx ubuntu", Input{"1.2.3.12", 80, "HTTP/1.1 200 OK\r\nServer: nginx/1.18.0 (Ubuntu)"},
			Result{"1.2.3.12", 80, "HTTP", "nginx", "1.18.0", "Ubuntu", 0.9}},
		{"http apache ubuntu", Input{"1.2.3.13", 443, "HTTP/1.1 200 OK\r\nServer: Apache/2.4.41 (Ubuntu)"},
			Result{"1.2.3.13", 443, "HTTP", "Apache", "2.4.41", "Ubuntu", 0.9}},
		{"mysql handshake 5.7.42", Input{"1.2.3.14", 3306, string([]byte{'J', 0, 0, 0, '\n', '5', '.', '7', '.', '4', '2', 0})},
			Result{"1.2.3.14", 3306, "MySQL", "MySQL", "5.7.42", "", 0.9}},
		{"redis pong", Input{"1.2.3.15", 6379, "+PONG"},
			Result{"1.2.3.15", 6379, "Redis", "Redis", "", "", 0.7}},
		{"ftp vsftpd", Input{"1.2.3.16", 21, "220 (vsFTPd 3.0.5)"},
			Result{"1.2.3.16", 21, "FTP", "vsftpd", "3.0.5", "", 0.9}},
		{"http nginx 8443", Input{"1.2.3.17", 8443, "HTTP/1.1 200 OK\r\nServer: nginx/1.25.3"},
			Result{"1.2.3.17", 8443, "HTTP", "nginx", "1.25.3", "", 0.9}},
		{"ssh old", Input{"1.2.3.18", 22, "SSH-1.99-OpenSSH_4.3"},
			Result{"1.2.3.18", 22, "SSH", "OpenSSH", "4.3", "", 0.95}},
		{"tls handshake", Input{"1.2.3.19", 9999, string([]byte{0x16, 0x03, 0x01, 0x00, 0xa5, 0x01, 0x00, 0x00, 0xa1})},
			Result{"1.2.3.19", 9999, "TLS", "", "", "", 0.6}},
		{"http iis", Input{"1.2.3.20", 8888, "HTTP/1.1 200 OK\r\nServer: Microsoft-IIS/10.0"},
			Result{"1.2.3.20", 8888, "HTTP", "Microsoft-IIS", "10.0", "", 0.9}},
		{"redis noauth", Input{"1.2.3.21", 6379, "-NOAUTH Authentication required."},
			Result{"1.2.3.21", 6379, "Redis", "Redis", "", "", 0.7}},
		{"ftp pureftpd", Input{"1.2.3.22", 21, "220 Welcome to Pure-FTPd"},
			Result{"1.2.3.22", 21, "FTP", "Pure-FTPd", "", "", 0.85}},
		{"garbage", Input{"1.2.3.23", 12345, "QUIT\r\n"},
			Result{"1.2.3.23", 12345, "unknown", "", "", "", 0}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := e.Identify([]Input{c.in})[0]
			if got != c.want {
				t.Errorf("identify %q:\n got  %+v\n want %+v", c.name, got, c.want)
			}
		})
	}
}
