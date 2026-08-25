package web

const stylesheet = `
:root { color-scheme: dark; font-family: Inter, Segoe UI, sans-serif; background: #101417; color: #edf2f4; }
* { box-sizing: border-box; }
body { margin: 0; min-height: 100vh; background: #101417; }
header { height: 58px; display: flex; align-items: center; justify-content: space-between; border-bottom: 1px solid #313a40; padding: 0 24px; background: #171c20; }
header strong { color: #ffcc4d; font-size: 18px; }
nav { display: flex; gap: 8px; }
nav a { color: #c9d2d8; text-decoration: none; padding: 8px 10px; border-radius: 4px; }
nav a:hover { background: #293137; color: #fff; }
main { padding: 24px; max-width: 1180px; margin: 0 auto; }
h1 { font-size: 26px; margin: 0 0 6px; }
.subtitle { color: #9eabb3; margin: 0 0 20px; }
.grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 12px; }
.panel { border: 1px solid #354149; background: #192126; border-radius: 6px; padding: 16px; min-height: 145px; }
.panel h2 { font-size: 14px; text-transform: uppercase; color: #9eabb3; margin: 0 0 12px; }
.value { font-size: 24px; font-variant-numeric: tabular-nums; }
.ok { color: #57d38c; }
.warn { color: #ffcc4d; }
.bad { color: #ff6b6b; }
button { border: 1px solid #697981; background: #263138; color: #fff; padding: 9px 12px; border-radius: 4px; cursor: pointer; }
button.primary { background: #e0a800; border-color: #e0a800; color: #151515; }
button:disabled { opacity: .45; cursor: not-allowed; }
input, select { width: 100%; background: #11171a; border: 1px solid #46535a; border-radius: 4px; color: #fff; padding: 9px; margin: 6px 0 10px; }
pre { margin: 0; white-space: pre-wrap; word-break: break-word; color: #c9d2d8; font-size: 12px; }
.toolbar { display: flex; flex-wrap: wrap; gap: 8px; margin: 16px 0; }
.statusline { border-left: 3px solid #e0a800; padding-left: 10px; margin-top: 12px; min-height: 22px; }
@media (max-width: 700px) { header { height: auto; padding: 12px; align-items: flex-start; } nav { flex-wrap: wrap; justify-content: flex-end; } main { padding: 16px; } }
`

const script = `
const endpoint = document.body.dataset.endpoint;
const output = document.querySelector('#output');
const statusline = document.querySelector('#statusline');
async function load() {
  const response = await fetch(endpoint, {headers: {'Accept':'application/json'}});
  const body = await response.json();
  output.textContent = JSON.stringify(body, null, 2);
  statusline.textContent = response.ok ? 'Live state received' : 'Request failed';
  statusline.className = 'statusline ' + (response.ok ? 'ok' : 'bad');
}
async function send(path, payload) {
  statusline.textContent = 'Submitting operation';
  const response = await fetch(path, {method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify(payload)});
  const body = await response.json();
  statusline.textContent = (response.ok ? 'Accepted: ' : 'Rejected: ') + (body.message || body.code || response.status);
  statusline.className = 'statusline ' + (response.ok ? 'ok' : 'bad');
  await load();
}
document.querySelectorAll('[data-action]').forEach(button => button.addEventListener('click', () => {
  const action = button.dataset.action;
  if (action === 'resume') send('/api/countdown/resume', {operation_id: crypto.randomUUID(), stable_seconds: 20});
  if (action === 'fill') send('/api/propellant/start', {operation_id: crypto.randomUUID(), kind: button.dataset.kind, arm: button.dataset.arm});
  if (action === 'arm') send('/api/umbilical/arm', {});
  if (action === 'hold') send('/api/holds', {source:'operator', reason:'manual safety hold'});
}));
load();
setInterval(load, 5000);
`
