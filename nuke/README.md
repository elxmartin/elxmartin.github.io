
<div align="center">
    <img src="nuke.svg" style="height: 100px" />
    <h3>Nuke</h3>
    <p>Targeted Nuclei Automation Engine</p>
</div>

### Install

```bash
go install github.com/elxmartin/elxmartin.github.io/nuke/cmd/nuke@latest
```
### Configure

```bash
nuke config \
  -url https://YOUR-WORKER.workers.dev \
  -token YOUR_SECRET

# It saves here: ~/.config/nuke/config.toml
``` 

### Scan

```bash
nuke -d hackerone.com
```
```bash
cat domains.txt | nuke
```
```bash
curl -s "https://bbscope.com/api/v1/targets/wildcards?platform=ywh&type=bbp" | nuke
```

---

### Commands

```bash
nuke version
```

```bash
nuke config -url https://YOUR-WORKER.workers.dev -token YOUR_SECRET
```

```bash
nuke scan -d hackerone.com
```

```bash
cat domains.txt | nuke
```

```bash
cat domains.txt | nuke scan
```

```bash
cat domains.txt | nuke import
```

```bash
nuke summary
```

```bash
nuke logs
```

```bash
nuke results
```

```bash
nuke status RUN_ID
```

```bash
nuke rerun
```

```bash
nuke delete nuke/log/example.log
```

```bash
nuke delete nuke/results/example.jsonl
```