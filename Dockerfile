FROM scratch

COPY ./dist/external-dns-webhook-opnsense /usr/bin/external-dns-webhook-opnsense

USER 65534:65534

ENTRYPOINT ["external-dns-webhook-opnsense"]
