# syntax=docker/dockerfile-upstream:master-labs
FROM scratch

ARG BIN=./dist/external-dns-webhook-opnsense

COPY --chmod=777 ${DIST} /usr/bin/external-dns-webhook-opnsense

USER 65534:65534

ENTRYPOINT ["external-dns-webhook-opnsense"]
