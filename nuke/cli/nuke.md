```bash
curl -s "https://bbscope.com/api/v1/targets/wildcards?platform=ywh&type=bbp" | nuke
```
```bash
nuke -d google.com
```
```bash
cat domains.txt | nuke
```
```bash
nuke config \
  -url https://YOUR-WORKER.workers.dev \
  -token YOUR_SECRET

# It saves here: ~/.config/nuke/config.txt
```