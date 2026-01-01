#include <stdio.h>
#include <string.h>
#include <errno.h>
#include "lwip/sockets.h"
#include "lwip/netdb.h"
#include "api_client.h"
/* Use sdk-provided esp_http_client when available; otherwise expose a
 * stub that returns an error and prints an actionable message.
 */
/* Implement HTTPS using mbedTLS (provided by the esp8266-rtos-sdk). This
 * keeps us on the RTOS SDK and avoids pulling an external esp-http-client
 * package. The implementation below is minimal: it performs a TLS handshake
 * with NO certificate verification (insecure) and fetches the response, then
 * parses it with cJSON. For production you should configure CA verification
 * or pin the server certificate.
 */

#include <stdlib.h>
#include <unistd.h>
#include "mbedtls/ssl.h"
#include "mbedtls/error.h"
#include "mbedtls/ctr_drbg.h"
#include "mbedtls/entropy.h"
#ifndef MBEDTLS_ERR_NET_SEND_FAILED
#define MBEDTLS_ERR_NET_SEND_FAILED -0x6000
#endif
#ifndef MBEDTLS_ERR_NET_RECV_FAILED
#define MBEDTLS_ERR_NET_RECV_FAILED -0x6001
#endif

/* JSON parsing removed for now — keep the response body raw and print it.
 * We'll add a small JSON parser later when application logic is ready.
 */

static int mbed_send(void *ctx, const unsigned char *buf, size_t len) {
    int fd = *(int *)ctx;
    int ret = send(fd, buf, len, 0);
    if (ret < 0) return MBEDTLS_ERR_NET_SEND_FAILED;
    return ret;
}

static int mbed_recv(void *ctx, unsigned char *buf, size_t len) {
    int fd = *(int *)ctx;
    int ret = recv(fd, buf, len, 0);
    if (ret < 0) return MBEDTLS_ERR_NET_RECV_FAILED;
    if (ret == 0) return MBEDTLS_ERR_SSL_PEER_CLOSE_NOTIFY;
    return ret;
}

static int parse_https_url(const char *url, char *host, size_t host_len, char *path, size_t path_len) {
    if (!url || !host || !path) return -1;
    const char *p = url;
    if (strncmp(p, "https://", 8) == 0) p += 8;
    else if (strncmp(p, "http://", 7) == 0) p += 7; /* allow fallback */

    const char *slash = strchr(p, '/');
    if (slash) {
        size_t hlen = (size_t)(slash - p);
        if (hlen >= host_len) return -1;
        memcpy(host, p, hlen);
        host[hlen] = '\0';
        strncpy(path, slash, path_len - 1);
        path[path_len - 1] = '\0';
    } else {
        if (strlen(p) >= host_len) return -1;
        strcpy(host, p);
        strncpy(path, "/", path_len - 1);
        path[path_len - 1] = '\0';
    }
    return 0;
}

int api_client_get_https(const char *url) {
    if (!url) return -1;

    char host[128];
    char path[256];
    if (parse_https_url(url, host, sizeof(host), path, sizeof(path)) != 0) {
        printf("api_client_get_https: parse URL failed\n");
        return -2;
    }

    /* DNS */
    struct hostent *he = gethostbyname(host);
    if (!he) {
        printf("api_client_get_https: DNS lookup failed for %s\n", host);
        return -3;
    }

    int sock = socket(AF_INET, SOCK_STREAM, 0);
    if (sock < 0) {
        printf("api_client_get_https: socket() failed\n");
        return -4;
    }

    struct sockaddr_in addr;
    memset(&addr, 0, sizeof(addr));
    addr.sin_family = AF_INET;
    addr.sin_port = htons(443);
    addr.sin_addr.s_addr = *((uint32_t *)he->h_addr);

    if (connect(sock, (struct sockaddr *)&addr, sizeof(addr)) != 0) {
        printf("api_client_get_https: connect() failed\n");
        close(sock);
        return -5;
    }

    /* mbedTLS setup */
    mbedtls_ssl_context ssl;
    mbedtls_ssl_config conf;
    mbedtls_ctr_drbg_context ctr_drbg;
    mbedtls_entropy_context entropy;
    const char *pers = "esp_client";

    mbedtls_ssl_init(&ssl);
    mbedtls_ssl_config_init(&conf);
    mbedtls_ctr_drbg_init(&ctr_drbg);
    mbedtls_entropy_init(&entropy);

    if (mbedtls_ctr_drbg_seed(&ctr_drbg, mbedtls_entropy_func, &entropy,
                              (const unsigned char *)pers, strlen(pers)) != 0) {
        printf("api_client_get_https: ctr_drbg_seed failed\n");
        goto cleanup_socket;
    }

    if (mbedtls_ssl_config_defaults(&conf,
            MBEDTLS_SSL_IS_CLIENT,
            MBEDTLS_SSL_TRANSPORT_STREAM,
            MBEDTLS_SSL_PRESET_DEFAULT) != 0) {
        printf("api_client_get_https: ssl_config_defaults failed\n");
        goto cleanup_drbg;
    }

    /* Insecure: do not verify server cert. For production set proper CA. */
    mbedtls_ssl_conf_authmode(&conf, MBEDTLS_SSL_VERIFY_NONE);
    mbedtls_ssl_conf_rng(&conf, mbedtls_ctr_drbg_random, &ctr_drbg);

    if (mbedtls_ssl_setup(&ssl, &conf) != 0) {
        printf("api_client_get_https: ssl_setup failed\n");
        goto cleanup_drbg;
    }

    if (mbedtls_ssl_set_hostname(&ssl, host) != 0) {
        /* not fatal on some builds */
    }

    mbedtls_ssl_set_bio(&ssl, &sock, mbed_send, mbed_recv, NULL);

    int ret;
    while ((ret = mbedtls_ssl_handshake(&ssl)) != 0) {
        if (ret != MBEDTLS_ERR_SSL_WANT_READ && ret != MBEDTLS_ERR_SSL_WANT_WRITE) {
            char errbuf[200];
            mbedtls_strerror(ret, errbuf, sizeof(errbuf));
            printf("api_client_get_https: handshake failed: %s\n", errbuf);
            goto cleanup_ssl;
        }
    }

    /* Build and send HTTP request */
    char req[512];
    int req_len = snprintf(req, sizeof(req),
        "GET %s HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n",
        path, host);

    if (req_len <= 0 || req_len >= (int)sizeof(req)) {
        printf("api_client_get_https: request too long\n");
        goto cleanup_ssl;
    }

    ret = mbedtls_ssl_write(&ssl, (const unsigned char *)req, req_len);
    if (ret <= 0) {
        printf("api_client_get_https: ssl_write failed: %d\n", ret);
        goto cleanup_ssl;
    }

    /* Read response */
    const int BUF_SIZE = 2048;
    char *buf = malloc(BUF_SIZE);
    if (!buf) {
        printf("api_client_get_https: malloc failed\n");
        goto cleanup_ssl;
    }
    int total = 0;
    for (;;) {
        ret = mbedtls_ssl_read(&ssl, (unsigned char *)buf + total, BUF_SIZE - 1 - total);
        if (ret == MBEDTLS_ERR_SSL_WANT_READ || ret == MBEDTLS_ERR_SSL_WANT_WRITE) continue;
        if (ret <= 0) break;
        total += ret;
        if (total >= BUF_SIZE - 1) break;
    }
    buf[total] = '\0';

    printf("api_client_get_https: received %d bytes:\n%s\n", total, buf);

    /* Find start of JSON body by locating first '{' and print raw JSON for now */
    char *json = strchr(buf, '{');
    if (json) {
        printf("api_client_get_https: JSON body:\n%s\n", json);
    } else {
        printf("api_client_get_https: no JSON body found\n");
    }

    free(buf);

cleanup_ssl:
    mbedtls_ssl_close_notify(&ssl);
    mbedtls_ssl_free(&ssl);
    mbedtls_ssl_config_free(&conf);
cleanup_drbg:
    mbedtls_ctr_drbg_free(&ctr_drbg);
    mbedtls_entropy_free(&entropy);
cleanup_socket:
    close(sock);
    return 0;
}

#define RECV_BUF_SIZE 1024

int api_client_get(const char *host, const char *path) {
    if (!host || !path) return -1;

    struct hostent *he = gethostbyname(host);
    if (!he) {
        printf("api_client_get: DNS lookup failed for %s\n", host);
        return -2;
    }

    int sock = socket(AF_INET, SOCK_STREAM, 0);
    if (sock < 0) {
        printf("api_client_get: socket() failed: %d\n", errno);
        return -3;
    }

    struct sockaddr_in addr;
    memset(&addr, 0, sizeof(addr));
    addr.sin_family = AF_INET;
    addr.sin_port = htons(80);
    addr.sin_addr.s_addr = *((uint32_t *)he->h_addr);

    if (connect(sock, (struct sockaddr *)&addr, sizeof(addr)) != 0) {
        printf("api_client_get: connect() failed: %d\n", errno);
        close(sock);
        return -4;
    }

    char req[512];
    int req_len = snprintf(req, sizeof(req),
        "GET %s HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n",
        path, host);

    if (req_len <= 0 || req_len >= (int)sizeof(req)) {
        printf("api_client_get: request too long\n");
        close(sock);
        return -5;
    }

    int sent = send(sock, req, req_len, 0);
    if (sent != req_len) {
        printf("api_client_get: send() sent %d of %d bytes\n", sent, req_len);
        close(sock);
        return -6;
    }

    char buf[RECV_BUF_SIZE];
    int r;
    printf("api_client_get: response from %s%s:\n", host, path);
    while ((r = recv(sock, buf, sizeof(buf) - 1, 0)) > 0) {
        buf[r] = '\0';
        printf("%s", buf);
    }

    if (r < 0) {
        printf("api_client_get: recv() failed: %d\n", errno);
    }

    printf("\n--- end response ---\n");
    close(sock);
    return 0;
}
