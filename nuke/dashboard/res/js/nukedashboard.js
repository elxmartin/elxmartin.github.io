const $ = id => document.getElementById(id);

let CONFIG = {
  worker_url: "",
  api_mode: "none"
};

let STATE = {
  summary: null,
  logs: [],
  results: []
};

window.onload = loadAll;

function repoBase(){
  return new URL("../", location.href).href;
}

async function loadConfig(){
  try{
    const res = await fetch("../config.json?cache=" + Date.now());
    if(res.ok) CONFIG = await res.json();
  }catch(e){}
}

async function loadAll(){
  await loadConfig();
  await Promise.all([
    loadSummary(),
    loadLogs(),
    loadResults()
  ]);
}

function showTab(name, btn){
  for(const id of ["scan","history","results","logs"]){
    $("tab-" + id).classList.add("hidden");
  }
  $("tab-" + name).classList.remove("hidden");

  document.querySelectorAll(".tab").forEach(t => t.classList.remove("active"));
  btn.classList.add("active");
}

async function api(path, options = {}){
  if(!CONFIG.worker_url){
    throw new Error("Missing worker_url in nuke/config.json");
  }

  const base = CONFIG.worker_url.replace(/\/$/, "");
  const headers = { ...(options.headers || {}) };

  const token = CONFIG.api_secret || localStorage.getItem("NUKE_API_KEY") || "";
  if(token) headers["Authorization"] = "Bearer " + token;

  const res = await fetch(base + path, {
    ...options,
    headers
  });

  const text = await res.text();
  try { return JSON.parse(text); }
  catch { return text; }
}

async function startScan(){
  const domains = $("domains").value.trim();
  if(!domains) return alert("Add domains first");

  try{
    const out = await api("/scan", {
      method: "POST",
      body: domains,
      headers: {"Content-Type":"text/plain"}
    });

    $("viewerTitle").textContent = "Scan Submitted";
    $("viewer").textContent = JSON.stringify(out, null, 2);
    alert("Scan submitted");
  }catch(e){
    alert(e.message);
  }
}

async function rerun(){
  try{
    const out = await api("/rerun", {
      method:"POST",
      body:"{}",
      headers:{"Content-Type":"application/json"}
    });

    $("viewerTitle").textContent = "Rerun Submitted";
    $("viewer").textContent = JSON.stringify(out, null, 2);
  }catch(e){
    alert(e.message);
  }
}

function importFile(){
  const f = $("fileInput").files[0];
  if(!f) return alert("Choose file first");

  const reader = new FileReader();
  reader.onload = () => $("domains").value = reader.result;
  reader.readAsText(f);
}

async function loadSummary(){
  try{
    const res = await fetch("../results/latest-summary.json?cache=" + Date.now());
    if(!res.ok) throw new Error("summary missing");

    const data = await res.json();
    STATE.summary = data;

    renderSummary(data);
    renderHistory(data);
  }catch(e){
    $("statStatus").textContent = "-";
    $("statRun").textContent = "-";
    $("statTechs").textContent = "0";
    $("statTemplates").textContent = "0";
    $("statFindings").textContent = "0";
    $("latestTime").textContent = "No results/latest-summary.json found yet.";
    $("scanTable").innerHTML = "";
    $("historyTable").innerHTML = "";
  }
}

function renderSummary(data){
  const scans = data.scans || [];
  const techs = scans.length;
  const templates = scans.reduce((n, s) => n + ((s.templates || []).length), 0);

  $("statStatus").innerHTML = scans.every(s => s.exit_code === 0) ? '<span class="ok">OK</span>' : '<span class="bad">ERR</span>';
  $("statRun").textContent = data.run_id || "-";
  $("statTechs").textContent = techs;
  $("statTemplates").textContent = templates;
  $("statFindings").textContent = "0";
  $("latestTime").textContent = `Started: ${data.started || "-"} | Finished: ${data.finished || "-"}`;

  $("scanTable").innerHTML = scans.map(s => `
    <tr>
      <td>${esc(s.tech)}</td>
      <td>${esc(s.input)}</td>
      <td>${(s.templates || []).map(t => `<span class="pill">${esc(shortName(t))}</span>`).join("")}</td>
      <td>${fileButton(s.result, "View Result")}</td>
      <td>${fileButton(s.log, "View Log")}</td>
      <td>${s.exit_code === 0 ? '<span class="ok">OK</span>' : '<span class="bad">ERR '+esc(s.exit_code)+'</span>'}</td>
    </tr>
  `).join("");
}

function renderHistory(data){
  const files = [
    ...(data.scans || []).map(s => s.result),
    ...(data.scans || []).map(s => s.log)
  ].filter(Boolean);

  $("historyTable").innerHTML = `
    <tr>
      <td>${esc(data.run_id || "-")}</td>
      <td>${esc(data.started || "-")}</td>
      <td>${esc(data.finished || "-")}</td>
      <td>${esc((data.scans || []).length)}</td>
      <td>${files.length}</td>
    </tr>
  `;
}

async function loadLogs(){
  STATE.logs = await listDir("../log/");
  renderFileTable("logsTable", STATE.logs);
}

async function loadResults(){
  STATE.results = await listDir("../results/");
  renderFileTable("resultsTable", STATE.results);
}

async function listDir(path){
  try{
    const index = await fetch(path + "?cache=" + Date.now());
    if(!index.ok) return [];

    const html = await index.text();
    const doc = new DOMParser().parseFromString(html, "text/html");
    const links = [...doc.querySelectorAll("a")]
      .map(a => a.getAttribute("href"))
      .filter(Boolean)
      .filter(h => !h.startsWith("?"))
      .filter(h => h !== "../")
      .map(h => new URL(h, new URL(path, location.href)).pathname)
      .map(p => decodeURIComponent(p))
      .filter(p => !p.endsWith("/"));

    return [...new Set(links)];
  }catch(e){
    return [];
  }
}

function renderFileTable(id, files){
  if(!files.length){
    $(id).innerHTML = `<tr><td colspan="3" class="muted">No files found.</td></tr>`;
    return;
  }

  $(id).innerHTML = files.map(file => `
    <tr>
      <td>${esc(file)}</td>
      <td class="muted">-</td>
      <td>
        <button class="gray inline" onclick="viewFile('${escAttr(file)}')">View</button>
        <a class="pill" target="_blank" href="${escAttr(file)}">Open</a>
      </td>
    </tr>
  `).join("");
}

function fileButton(path, label){
  if(!path) return "-";
  const href = "../" + path.replace(/^nuke\//, "");
  return `<button class="gray inline" onclick="viewFile('${escAttr(href)}')">${esc(label)}</button>`;
}

async function viewFile(path){
  try{
    $("viewerTitle").textContent = path;
    const res = await fetch(path + "?cache=" + Date.now());
    $("viewer").textContent = await res.text();
  }catch(e){
    $("viewer").textContent = "Could not load file.";
  }
}

function clearViewer(){
  $("viewerTitle").textContent = "Viewer";
  $("viewer").textContent = "Select a log or result file.";
}

function shortName(path){
  return String(path || "").split("/").pop();
}

function esc(s){
  return String(s ?? "")
    .replaceAll("&","&amp;")
    .replaceAll("<","&lt;")
    .replaceAll(">","&gt;")
    .replaceAll('"',"&quot;");
}

function escAttr(s){
  return esc(s).replaceAll("'","&#39;");
}