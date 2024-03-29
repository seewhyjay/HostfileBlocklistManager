# HostfileBlocklistManager
A simple command-line based DNS blocklist manager.

## Features
* Use blocklists compatible with [Pi-hole](https://github.com/pi-hole/pi-hole) or [AdGuard Home](https://github.com/AdguardTeam/AdGuardHome)
* Whitelist domains (Domains whitelisted take precedence over blocklists and blacklist)
* Blacklist domains

## Usage
### Add a blocklist
**NOTE: You need to run this within an elevated Command Prompt/Powershell because of the need to edit the `hosts` file**
```
.\HostfileBlocklistManager.exe -blocklist=<URL>
```
### Whitelist a domain
```
.\HostfileBlocklistManager.exe -whitelist=<domain>
```
### Blacklist a domain
```
.\HostfileBlocklistManager.exe -blacklist=<domain>
```
### Update all blocklists
**NOTE: You need to run this within an elevated Command Prompt/Powershell because of the need to edit the `hosts` file**
```
.\HostfileBlocklistManager.exe -update
```

# Examples
**NOTE: You need to run this within an elevated Command Prompt/Powershell because of the need to edit the `hosts` file**
```
.\HostfileBlocklistManager.exe -blocklist=https://adaway.org/hosts.txt
```

## Planned features
* [ ] Deduplicate entries in blocklists
* [ ] MacOS support
