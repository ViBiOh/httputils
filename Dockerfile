FROM scratch

HEALTHCHECK --retries=10 CMD http://localhost/health

COPY bin/alcotest /bin/sh
