# Nuke

Targeted Nuclei Automation Engine

## Install

```bash
go install github.com/elxmartin/nuke/cmd/nuke@latest
```

## Configure

```bash
nuke config \
-url https://nuke.elxmartin.workers.dev \
-token XXXXX
```

## Scan

```bash
echo hackerone.com | nuke
```

```bash
nuke -d hackerone.com
```
