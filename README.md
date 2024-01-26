# HostfileBlocklistManager
A simple command-line based DNS blocklist manager.

## Usage
To add a blocklist use,
```
.\HostfileBlocklistManager.exe -blocklist=<URL>
```

# Example
**NOTE: You need to run this within an elevated Command Prompt/Powershell because of the need to edit the `hosts` file**
```
.\HostfileBlocklistManager.exe -blocklist=https://adaway.org/hosts.txt
```