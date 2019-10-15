# Marathon - Consul

Marathon-Consul registers [Marathon](https://mesosphere.github.io/marathon/)'s services into [Consul](https://www.consul.io/), to be later used by [FabioLB](https://fabiolb.net/).

It only registers Marathon's service that has HTTP health check (as required by Fabio).

It also looks for specific Marathon's label (`urlprefix`) and convert it as Consul tags.
Checkout [Fabio's Quick Start](https://fabiolb.net/quickstart/) for `urlprefix` format.
Several `urlprefix` can be combined using semicolon (`;`). Additionally, multiple `urlprefix` can be added in labels by adding underscore followed by any number (eg: `urlprefix_1`).

Marathon example:
```
{
    ...,
    
    "labels": {
        "urlprefix": "example.com:443/; example.com:80/ redirect=301,https://example.com/$path",
        "urlprefix_2": "example2.com:443/; example2.com:80/ redirect=301,https://example2.com/$path",
    },
    ...
}
```

## Additional Tools

Additional tools in this repo is `certconsul`.

It scans a directory for certificates (.crt & .key) and sync it into Consul's Key Value stores.
