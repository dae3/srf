#ifndef API_CLIENT_H
#define API_CLIENT_H

/* Simple blocking HTTP GET helper using lwIP sockets.
 * This is intentionally minimal: it connects to host:80 and performs
 * a plain HTTP/1.1 GET. It prints the response to stdout and returns
 * 0 on success or negative on error.
 *
 * NOTE: This helper does NOT perform TLS. For production HTTPS you
 * should replace this with an HTTPS-capable client (esp_http_client
 * or a mbedTLS-backed socket flow).
 */

int api_client_get(const char *host, const char *path);
int api_client_get_https(const char *url);

#endif // API_CLIENT_H
