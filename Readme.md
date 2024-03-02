### Proxy-Collector

This collects Free Proxies/VPNs and Cloudflare Edgetunnel Domains for neobox,
and parses version list for version-manager.
And also restores the collected data into a github/gitee repository.

### Usage
```bash
go install github.com/moqsien/proxy-collector/pkgs/pxy
```

### Commands
```bash
mq@mqMac pxy % ./pxy -h
Usage:
   [command]

Proxy Collector Commands:
  add-domain          Adds rawDomains to rawDomain list.
  add-subscribedUrls  Adds urls to subscribedUrl list.
  get-proxies         Collects proxies.
  reset-cryptokey     Resets cryptoKey.
  set-localproxy      Sets local proxy for fetcher.
  show-cryptokey      Shows cryptoKey.
  show-rawdomains     Shows rawDomain list.
  show-subscribedurls Shows subscribed urls.
  test-domains        Tests domains for edgetunnels.
  version-add-repo    Add github repos for parsing release list.
  version-fetch       Get version list for gvc.

Additional Commands:
  completion          Generate the autocompletion script for the specified shell
  help                Help about any command

Flags:
  -h, --help   help for this command

Use " [command] --help" for more information about a command.
```
