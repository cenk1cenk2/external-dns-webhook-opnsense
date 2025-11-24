# syntax=docker/dockerfile-upstream:master-labs
FROM scratch

COPY --chmod=777 ./dist/external-dns-webhook-opnsense /usr/bin/external-dns-webhook-opnsense

USER 65534:65534

ENTRYPOINT ["external-dns-webhook-opnsense"]
