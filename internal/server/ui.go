package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashHTML))
}

const dashHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>Outpost</title>
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;700&display=swap" rel="stylesheet">
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c94444;--orange:#d4843a;--blue:#5b8dd9;--mono:'JetBrains Mono',monospace}
*{margin:0;padding:0;box-sizing:border-box}
body{background:var(--bg);color:var(--cream);font-family:var(--mono);line-height:1.5;font-size:13px}
.hdr{padding:.8rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center;gap:1rem;flex-wrap:wrap}
.hdr h1{font-size:.9rem;letter-spacing:2px}
.hdr h1 span{color:var(--rust)}
.main{padding:1.2rem 1.5rem;max-width:1200px;margin:0 auto}
.stats{display:grid;grid-template-columns:repeat(5,1fr);gap:.5rem;margin-bottom:1rem}
.st{background:var(--bg2);border:1px solid var(--bg3);padding:.7rem;text-align:center}
.st-v{font-size:1.2rem;font-weight:700;color:var(--gold)}
.st-v.green{color:var(--green)}
.st-v.red{color:var(--red)}
.st-l{font-size:.5rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-top:.2rem}
.toolbar{display:flex;gap:.5rem;margin-bottom:1rem;flex-wrap:wrap;align-items:center}
.search{flex:1;min-width:180px;padding:.4rem .6rem;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.search:focus{outline:none;border-color:var(--leather)}
.filter-sel{padding:.4rem .5rem;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.65rem}

.grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(320px,1fr));gap:.6rem}
.card{background:var(--bg2);border:1px solid var(--bg3);padding:.9rem 1rem;display:flex;flex-direction:column;gap:.4rem;cursor:pointer;transition:border-color .15s}
.card:hover{border-color:var(--leather)}
.card.offline{opacity:.6}
.card.online{border-left:3px solid var(--green)}
.card.offline{border-left:3px solid var(--cm)}
.card.unknown{border-left:3px solid var(--orange)}
.card-top{display:flex;justify-content:space-between;align-items:flex-start;gap:.5rem}
.card-name{font-size:.85rem;font-weight:700;color:var(--cream)}
.card-host{font-size:.6rem;color:var(--cd);margin-top:.1rem;font-family:var(--mono)}
.card-meta{font-size:.55rem;color:var(--cm);display:flex;gap:.5rem;flex-wrap:wrap;margin-top:.2rem}
.badge{font-size:.5rem;padding:.12rem .35rem;text-transform:uppercase;letter-spacing:1px;border:1px solid var(--bg3);color:var(--cm);font-weight:700}
.badge.online{border-color:var(--green);color:var(--green)}
.badge.offline{border-color:var(--cm);color:var(--cm)}
.badge.unknown{border-color:var(--orange);color:var(--orange)}

.metrics{display:flex;flex-direction:column;gap:.3rem;margin-top:.4rem}
.metric{display:flex;align-items:center;gap:.5rem}
.metric-label{font-size:.5rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;width:36px;flex-shrink:0}
.metric-bar{flex:1;height:6px;background:var(--bg3);position:relative;overflow:hidden}
.metric-bar-fill{position:absolute;top:0;left:0;bottom:0;background:var(--green);transition:width .3s}
.metric-bar-fill.warn{background:var(--orange)}
.metric-bar-fill.crit{background:var(--red)}
.metric-pct{font-family:var(--mono);font-size:.6rem;color:var(--cd);min-width:36px;text-align:right}
.uptime{font-size:.55rem;color:var(--cm);font-style:italic;margin-top:.2rem}
.card-extra{font-size:.55rem;color:var(--cd);margin-top:.4rem;padding-top:.3rem;border-top:1px dashed var(--bg3);display:flex;flex-direction:column;gap:.15rem}
.card-extra-row{display:flex;gap:.4rem}
.card-extra-label{color:var(--cm);text-transform:uppercase;letter-spacing:.5px;min-width:90px}
.card-extra-val{color:var(--cream)}

.btn{font-family:var(--mono);font-size:.6rem;padding:.3rem .55rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd);transition:.15s}
.btn:hover{border-color:var(--leather);color:var(--cream)}
.btn-p{background:var(--rust);border-color:var(--rust);color:#fff}
.btn-p:hover{opacity:.85;color:#fff}
.btn-sm{font-size:.55rem;padding:.2rem .4rem}
.btn-del{color:var(--red);border-color:#3a1a1a}
.btn-del:hover{border-color:var(--red);color:var(--red)}

.modal-bg{display:none;position:fixed;inset:0;background:rgba(0,0,0,.65);z-index:100;align-items:center;justify-content:center}
.modal-bg.open{display:flex}
.modal{background:var(--bg2);border:1px solid var(--bg3);padding:1.5rem;width:480px;max-width:92vw;max-height:90vh;overflow-y:auto}
.modal h2{font-size:.8rem;margin-bottom:1rem;color:var(--rust);letter-spacing:1px}
.fr{margin-bottom:.6rem}
.fr label{display:block;font-size:.55rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-bottom:.2rem}
.fr input,.fr select,.fr textarea{width:100%;padding:.4rem .5rem;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.fr input:focus,.fr select:focus,.fr textarea:focus{outline:none;border-color:var(--leather)}
.row2{display:grid;grid-template-columns:1fr 1fr;gap:.5rem}
.fr-section{margin-top:1rem;padding-top:.8rem;border-top:1px solid var(--bg3)}
.fr-section-label{font-size:.55rem;color:var(--rust);text-transform:uppercase;letter-spacing:1px;margin-bottom:.5rem}
.acts{display:flex;gap:.4rem;justify-content:flex-end;margin-top:1rem}
.acts .btn-del{margin-right:auto}
.empty{text-align:center;padding:3rem;color:var(--cm);font-style:italic;font-size:.85rem}
.agent-help{margin-top:1rem;padding:1rem;background:var(--bg2);border:1px solid var(--bg3);font-size:.65rem;color:var(--cd)}
.agent-help code{background:var(--bg);padding:.1rem .3rem;color:var(--gold)}
@media(max-width:600px){.stats{grid-template-columns:repeat(2,1fr)}}
</style>
</head>
<body>

<div class="hdr">
<h1 id="dash-title"><span>&#9670;</span> OUTPOST</h1>
<button class="btn btn-p" onclick="openNew()">+ Register Host</button>
</div>

<div class="main">
<div class="stats" id="stats"></div>
<div class="toolbar">
<input class="search" id="search" placeholder="Search hostname, ip, os..." oninput="debouncedRender()">
<select class="filter-sel" id="status-filter" onchange="render()">
<option value="">All Statuses</option>
<option value="online">Online</option>
<option value="offline">Offline</option>
<option value="unknown">Unknown</option>
</select>
</div>
<div id="grid" class="grid"></div>

<div class="agent-help">
<strong>Agent reporting:</strong> Have your hosts POST metrics to <code>/api/report</code> every 60 seconds with JSON like:
<code>{"hostname":"db-01","cpu_pct":42.5,"mem_pct":68.0,"disk_pct":31.2,"uptime":"3d 14h","ip":"10.0.1.5","os":"Ubuntu 22.04"}</code>.
Hosts that haven't reported in 120 seconds are marked offline.
</div>
</div>

<div class="modal-bg" id="mbg" onclick="if(event.target===this)closeModal()">
<div class="modal" id="mdl"></div>
</div>

<script>
var A='/api';
var RESOURCE='hosts';

var fields=[
{name:'name',label:'Display Name',type:'text',required:true},
{name:'hostname',label:'Hostname',type:'text'},
{name:'ip',label:'IP Address',type:'text'},
{name:'os',label:'Operating System',type:'text'}
];

var hosts=[],hostExtras={},editId=null,searchTimer=null;

function fmtAgo(s){
if(!s)return'never';
try{
var d=new Date(s);
if(isNaN(d.getTime()))return s;
var diffMs=Date.now()-d;
if(diffMs<0)return'just now';
var sec=Math.floor(diffMs/1000);
if(sec<60)return sec+'s ago';
var min=Math.floor(sec/60);
if(min<60)return min+'m ago';
var hours=Math.floor(min/60);
if(hours<24)return hours+'h ago';
return Math.floor(hours/24)+'d ago';
}catch(e){return s}
}

function fieldByName(n){for(var i=0;i<fields.length;i++)if(fields[i].name===n)return fields[i];return null}

function debouncedRender(){
clearTimeout(searchTimer);
searchTimer=setTimeout(render,200);
}

async function load(){
try{
var resps=await Promise.all([
fetch(A+'/hosts').then(function(r){return r.json()}),
fetch(A+'/stats').then(function(r){return r.json()})
]);
hosts=resps[0].hosts||[];
renderStats(resps[1]||{});

try{
var ex=await fetch(A+'/extras/'+RESOURCE).then(function(r){return r.json()});
hostExtras=ex||{};
hosts.forEach(function(h){
var x=hostExtras[h.id];
if(!x)return;
Object.keys(x).forEach(function(k){if(h[k]===undefined)h[k]=x[k]});
});
}catch(e){hostExtras={}}
}catch(e){
console.error('load failed',e);
hosts=[];
}
render();
}

function renderStats(s){
var total=s.total||0;
var online=s.online||0;
var offline=s.offline||0;
var maxCpu=Math.round(s.max_cpu_pct||0);
var maxMem=Math.round(s.max_mem_pct||0);
document.getElementById('stats').innerHTML=
'<div class="st"><div class="st-v">'+total+'</div><div class="st-l">Hosts</div></div>'+
'<div class="st"><div class="st-v green">'+online+'</div><div class="st-l">Online</div></div>'+
'<div class="st"><div class="st-v red">'+offline+'</div><div class="st-l">Offline</div></div>'+
'<div class="st"><div class="st-v">'+maxCpu+'%</div><div class="st-l">Peak CPU</div></div>'+
'<div class="st"><div class="st-v">'+maxMem+'%</div><div class="st-l">Peak Mem</div></div>';
}

function render(){
var q=(document.getElementById('search').value||'').toLowerCase();
var sf=document.getElementById('status-filter').value;

var f=hosts.slice();
if(q)f=f.filter(function(h){
return(h.name||'').toLowerCase().includes(q)||
(h.hostname||'').toLowerCase().includes(q)||
(h.ip||'').toLowerCase().includes(q)||
(h.os||'').toLowerCase().includes(q);
});
if(sf)f=f.filter(function(h){return h.status===sf});

if(!f.length){
var msg=window._emptyMsg||'No hosts reporting yet.';
document.getElementById('grid').innerHTML='<div class="empty" style="grid-column:1/-1">'+esc(msg)+'</div>';
return;
}

var h='';
f.forEach(function(host){h+=cardHTML(host)});
document.getElementById('grid').innerHTML=h;
}

function metricBar(label,pct){
var p=parseFloat(pct||0);
if(p<0)p=0;if(p>100)p=100;
var cls='';
if(p>=90)cls='crit';else if(p>=75)cls='warn';
return '<div class="metric"><span class="metric-label">'+label+'</span>'+
'<div class="metric-bar"><div class="metric-bar-fill '+cls+'" style="width:'+p+'%"></div></div>'+
'<span class="metric-pct">'+p.toFixed(0)+'%</span></div>';
}

function cardHTML(host){
var status=host.status||'unknown';
var h='<div class="card '+esc(status)+'" onclick="openEdit(\''+esc(host.id)+'\')">';

h+='<div class="card-top">';
h+='<div style="flex:1;min-width:0">';
h+='<div class="card-name">'+esc(host.name)+'</div>';
if(host.hostname&&host.hostname!==host.name)h+='<div class="card-host">'+esc(host.hostname)+'</div>';
h+='</div>';
h+='<span class="badge '+esc(status)+'">'+esc(status)+'</span>';
h+='</div>';

if(host.ip||host.os){
h+='<div class="card-meta">';
if(host.ip)h+='<span>'+esc(host.ip)+'</span>';
if(host.os)h+='<span>'+esc(host.os)+'</span>';
h+='</div>';
}

h+='<div class="metrics">';
h+=metricBar('CPU',host.cpu_pct);
h+=metricBar('MEM',host.mem_pct);
h+=metricBar('DISK',host.disk_pct);
h+='</div>';

h+='<div class="uptime">';
if(host.uptime)h+='up '+esc(host.uptime)+' &middot; ';
h+='reported '+esc(fmtAgo(host.last_report));
h+='</div>';

// Custom field display
var customRows='';
fields.forEach(function(f){
if(!f.isCustom)return;
var v=host[f.name];
if(v===undefined||v===null||v==='')return;
customRows+='<div class="card-extra-row">';
customRows+='<span class="card-extra-label">'+esc(f.label)+'</span>';
customRows+='<span class="card-extra-val">'+esc(String(v))+'</span>';
customRows+='</div>';
});
if(customRows)h+='<div class="card-extra">'+customRows+'</div>';

h+='</div>';
return h;
}

// ─── Modal ────────────────────────────────────────────────────────

function fieldHTML(f,value){
var v=value;
if(v===undefined||v===null)v='';
var req=f.required?' *':'';
var ph=f.placeholder?(' placeholder="'+esc(f.placeholder)+'"'):'';
var h='<div class="fr"><label>'+esc(f.label)+req+'</label>';

if(f.type==='select'){
h+='<select id="f-'+f.name+'">';
if(!f.required)h+='<option value="">Select...</option>';
(f.options||[]).forEach(function(o){
var sel=(String(v)===String(o))?' selected':'';
h+='<option value="'+esc(String(o))+'"'+sel+'>'+esc(String(o))+'</option>';
});
h+='</select>';
}else if(f.type==='textarea'){
h+='<textarea id="f-'+f.name+'" rows="3"'+ph+'>'+esc(String(v))+'</textarea>';
}else if(f.type==='number'){
h+='<input type="number" id="f-'+f.name+'" value="'+esc(String(v))+'"'+ph+'>';
}else{
h+='<input type="text" id="f-'+f.name+'" value="'+esc(String(v))+'"'+ph+'>';
}
h+='</div>';
return h;
}

function formHTML(host){
var h0=host||{};
var isEdit=!!host;
var h='<h2>'+(isEdit?'EDIT HOST':'REGISTER HOST')+'</h2>';

if(!isEdit){
h+='<div style="font-size:.6rem;color:var(--cm);margin-bottom:.7rem;font-style:italic">Hosts auto-register when they POST to /api/report. Use this only for manual setup.</div>';
}

h+=fieldHTML(fieldByName('name'),h0.name);
h+=fieldHTML(fieldByName('hostname'),h0.hostname);
h+='<div class="row2">'+fieldHTML(fieldByName('ip'),h0.ip)+fieldHTML(fieldByName('os'),h0.os)+'</div>';

var customFields=fields.filter(function(f){return f.isCustom});
if(customFields.length){
var label=window._customSectionLabel||'Additional Details';
h+='<div class="fr-section"><div class="fr-section-label">'+esc(label)+'</div>';
customFields.forEach(function(f){h+=fieldHTML(f,h0[f.name])});
h+='</div>';
}

h+='<div class="acts">';
if(isEdit)h+='<button class="btn btn-del" onclick="delItem()">Delete</button>';
h+='<button class="btn" onclick="closeModal()">Cancel</button>';
h+='<button class="btn btn-p" onclick="submit()">'+(isEdit?'Save':'Register')+'</button>';
h+='</div>';
return h;
}

function openNew(){
editId=null;
document.getElementById('mdl').innerHTML=formHTML();
document.getElementById('mbg').classList.add('open');
var n=document.getElementById('f-name');if(n)n.focus();
}

function openEdit(id){
var h=null;
for(var i=0;i<hosts.length;i++){if(hosts[i].id===id){h=hosts[i];break}}
if(!h)return;
editId=id;
document.getElementById('mdl').innerHTML=formHTML(h);
document.getElementById('mbg').classList.add('open');
}

function closeModal(){
document.getElementById('mbg').classList.remove('open');
editId=null;
}

async function submit(){
var nameEl=document.getElementById('f-name');
if(!nameEl||!nameEl.value.trim()){alert('Display name is required');return}

var body={};
var extras={};
fields.forEach(function(f){
var el=document.getElementById('f-'+f.name);
if(!el)return;
var val=el.value.trim();
if(f.isCustom)extras[f.name]=val;
else body[f.name]=val;
});

var savedId=editId;
try{
if(editId){
var r1=await fetch(A+'/hosts/'+editId,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
if(!r1.ok){var e1=await r1.json().catch(function(){return{}});alert(e1.error||'Save failed');return}
}else{
var r2=await fetch(A+'/hosts',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});
if(!r2.ok){var e2=await r2.json().catch(function(){return{}});alert(e2.error||'Register failed');return}
var created=await r2.json();
savedId=created.id;
}
if(savedId&&Object.keys(extras).length){
await fetch(A+'/extras/'+RESOURCE+'/'+savedId,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(extras)}).catch(function(){});
}
}catch(e){alert('Network error: '+e.message);return}
closeModal();
load();
}

async function delItem(){
if(!editId)return;
if(!confirm('Delete this host?'))return;
await fetch(A+'/hosts/'+editId,{method:'DELETE'});
closeModal();
load();
}

function esc(s){
if(s===undefined||s===null)return'';
var d=document.createElement('div');
d.textContent=String(s);
return d.innerHTML;
}

document.addEventListener('keydown',function(e){if(e.key==='Escape')closeModal()});

// Auto-refresh every 15s for live metrics
setInterval(load,15000);

(function loadPersonalization(){
fetch('/api/config').then(function(r){return r.json()}).then(function(cfg){
if(!cfg||typeof cfg!=='object')return;

if(cfg.dashboard_title){
var h1=document.getElementById('dash-title');
if(h1)h1.innerHTML='<span>&#9670;</span> '+esc(cfg.dashboard_title);
document.title=cfg.dashboard_title;
}

if(cfg.empty_state_message)window._emptyMsg=cfg.empty_state_message;
if(cfg.primary_label)window._customSectionLabel=cfg.primary_label+' Details';

if(Array.isArray(cfg.custom_fields)){
cfg.custom_fields.forEach(function(cf){
if(!cf||!cf.name||!cf.label)return;
if(fieldByName(cf.name))return;
fields.push({
name:cf.name,
label:cf.label,
type:cf.type||'text',
options:cf.options||[],
isCustom:true
});
});
}
}).catch(function(){
}).finally(function(){
load();
});
})();
</script>
</body>
</html>`
