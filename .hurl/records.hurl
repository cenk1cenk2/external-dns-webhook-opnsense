GET http://localhost:8888/records
Accept: application/external.dns.webhook+json;version=1

POST http://localhost:8888/records
Content-Type: application/external.dns.webhook+json;version=1

{
  "create": [
    {
      "dnsName": "test.test.arpa",
      "targets": [
        "127.0.0.1"
      ],
      "recordType": "A"
    },
    {
      "dnsName": "something",
      "targets": [
        "test"
      ],
      "recordType": "TXT"
    }
  ]
}

POST http://localhost:8888/records
Content-Type: application/external.dns.webhook+json;version=1

{
  "delete": [
    {
      "dnsName": "test.test.arpa",
      "targets": [
        "127.0.0.1"
      ],
      "recordType": "A"
    },
    {
      "dnsName": "something",
      "targets": [
        "test"
      ],
      "recordType": "TXT"
    }
  ]
}
